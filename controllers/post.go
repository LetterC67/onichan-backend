package controllers

import (
	"fmt"
	"net/http"
	"onichan/database"
	"onichan/model"
	"onichan/services"
	"onichan/websocket"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Payload struct {
	Title        *string `json:"title"`
	Content      string  `json:"content" binding:"required"`
	IsMasterPost bool    `json:"is_master_post"`
	ParentPostID *uint   `json:"parent_post_id"`
	ReplyToID    *uint   `json:"reply_to_id"`
	CategoryID   uint    `json:"category_id" binding:"required"`
}

var pageSize int

func LoadPageSize() {
	_pageSize, err := strconv.Atoi(os.Getenv("PAGE_SIZE"))
	pageSize = _pageSize
	if err != nil {
		fmt.Println("PAGE_SIZE is not set")
	}
}

func validatePost(payload interface{}) (bool, string) {
	var isMasterPost bool
	var parentPostID, replyToID *uint
	var title *string
	var categoryID uint

	switch p := payload.(type) {
	case Payload:
		isMasterPost = p.IsMasterPost
		parentPostID = p.ParentPostID
		replyToID = p.ReplyToID
		title = p.Title
		categoryID = p.CategoryID
	case model.Post:
		isMasterPost = p.IsMasterPost
		parentPostID = p.ParentPostID
		replyToID = p.ReplyToID
		title = p.Title
		categoryID = p.CategoryID
	default:
		return false, "Invalid payload type"
	}

	if isMasterPost && parentPostID != nil {
		return false, "Master post must not have parent post"
	}

	if isMasterPost && replyToID != nil {
		return false, "Master post must not have reply to post"
	}

	if !isMasterPost && parentPostID == nil {
		return false, "Non-master post must have parent post"
	}

	if !isMasterPost && title != nil {
		return false, "Non-master post must not have title"
	}

	if isMasterPost && title == nil {
		return false, "Master post must have title"
	}

	if parentPostID != nil {
		var parentPost model.Post
		if err := database.Database.First(&parentPost, parentPostID).Error; err != nil {
			return false, "Parent post not found"
		}

		if parentPost.CategoryID != categoryID {
			return false, "Parent post must have the same category"
		}
	}

	if replyToID != nil {
		if err := database.Database.First(&model.Post{}, replyToID).Error; err != nil {
			return false, "Reply to post not found"
		}
	}

	if err := database.Database.First(&model.Category{}, categoryID).Error; err != nil {
		return false, "Category not found"
	}

	return true, ""
}

// CreatePost godoc
// @Summary      Create a new post
// @Description  Creates a master post or a reply. Master posts require a title and cannot have a parent, while replies must have a parent and must not have a title.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        payload  body      Payload  true  "Post payload"
// @Success      200      {object}  map[string]interface{}  "page, id"
// @Failure      400      {object}  map[string]interface{}  "Bad Request"
// @Failure      500      {object}  map[string]interface{}  "Internal Server Error"
// @Security     ApiKeyAuth
// @Router       /posts [post]
func CreatePost(c *gin.Context) {
	var payload Payload
	var parentPost model.Post
	var replyToPost model.Post

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ok, message := validatePost(payload); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	userID, _ := c.Get("user_id")
	userIDUint := uint(userID.(float64))

	user := model.User{}
	if err := database.Database.First(&user, userIDUint).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	post := model.Post{
		UserID:       userIDUint,
		User:         user,
		Title:        payload.Title,
		Content:      payload.Content,
		IsMasterPost: payload.IsMasterPost,
		ParentPostID: payload.ParentPostID,
		ReplyToID:    payload.ReplyToID,
		CategoryID:   payload.CategoryID,
		LastUpdated:  time.Now(),
	}

	if payload.ReplyToID != nil {
		if err := database.Database.First(&replyToPost, payload.ReplyToID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reply to post not found"})
			return
		}
	}

	if payload.ParentPostID != nil {
		if err := database.Database.First(&parentPost, payload.ParentPostID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent post not found"})
			return
		}

		if payload.ReplyToID != nil {
			if err := database.Database.First(&replyToPost, payload.ReplyToID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Reply to post not found"})
				return
			}
		}

		parentPost.LastUpdated = time.Now()
		if err := database.Database.Save(&parentPost).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if result := database.Database.Create(&post); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if payload.ParentPostID != nil {
		websocket.SendNewPostSignal(*post.ParentPostID, userIDUint)
	}

	if payload.ReplyToID != nil && replyToPost.UserID != userIDUint {
		err := services.CreateNotification(replyToPost.UserID, userIDUint, post.ID, "reply")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if payload.ParentPostID != nil && parentPost.UserID != userIDUint {
		err := services.CreateNotification(parentPost.UserID, userIDUint, post.ID, "comment")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	var count int64
	if err := database.Database.Model(&model.Post{}).
		Where("parent_post_id = ? AND created_at < ?", post.ParentPostID, post.CreatedAt).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page": (int(count) + pageSize + 1) / pageSize,
		"id":   post.ID,
	})
}

// ListPosts godoc
// @Summary      List posts
// @Description  Retrieves a paginated list of master posts from a category, identified by either category ID or category name.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        category_id    query     string  false  "Category ID"
// @Param        category_name  query     string  false  "Category Name"
// @Param        page           query     int     false  "Page number"  default(1)
// @Success      200  {object}  map[string]interface{}  "List of posts and total_pages"
// @Failure      400  {object}  map[string]interface{}  "Bad Request"
// @Failure      404  {object}  map[string]interface{}  "Not Found"
// @Failure      500  {object}  map[string]interface{}  "Internal Server Error"
// @Router       /posts [get]
func ListPosts(c *gin.Context) {
	var posts []model.Post
	var totalPosts int64
	categoryID := c.Query("category_id")
	categoryName := c.Query("category_name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	if categoryID == "" && categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}

	if categoryName != "" {
		var category model.Category
		if err := database.Database.Where("name = ?", categoryName).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		categoryID = strconv.Itoa(int(category.ID))
	}

	offset := (page - 1) * pageSize

	if err := database.Database.
		Preload("User").
		Order("last_updated DESC").
		Where("category_id = ? AND is_master_post = ?", categoryID, true).
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.
		Model(&model.Post{}).
		Where("category_id = ? AND is_master_post = ?", categoryID, true).
		Count(&totalPosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range posts {
		var replyCount int64
		database.Database.Model(&model.Post{}).
			Where("parent_post_id = ?", posts[i].ID).
			Count(&replyCount)
		posts[i].RepliesCount = int(replyCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"total_pages": (int(totalPosts) + pageSize - 1) / pageSize,
	})
}

func getPostReactions(postID uint, c *gin.Context) ([]model.PostReactionCount, []model.PostReactionCount, error) {
	userID := c.Query("user_id")

	var reactions []model.Reaction
	if err := database.Database.Find(&reactions).Error; err != nil {
		return nil, nil, err
	}

	reactionsCount := make([]model.PostReactionCount, 0)
	userReactionsCount := make([]model.PostReactionCount, 0)

	for _, reaction := range reactions {
		var reactionCount, userReactionCount int64

		if userID != "" {
			database.Database.Model(&model.PostReaction{}).
				Where("post_id = ? AND user_id = ? AND reaction_id = ?", postID, userID, reaction.ID).
				Count(&userReactionCount)
		}

		database.Database.Model(&model.PostReaction{}).
			Where("post_id = ? AND reaction_id = ?", postID, reaction.ID).
			Count(&reactionCount)

		reactionsCount = append(reactionsCount, model.PostReactionCount{
			Reaction: reaction,
			Count:    int(reactionCount),
		})

		if userReactionCount != 0 {
			userReactionsCount = append(userReactionsCount, model.PostReactionCount{
				Reaction: reaction,
				Count:    int(userReactionCount),
			})
		}
	}

	return reactionsCount, userReactionsCount, nil
}

// GetPost godoc
// @Summary      Get a post and its replies
// @Description  Retrieves a specific post by its ID. Also returns any replies, category and user details, reaction data, etc. Pagination is applied to replies.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id    path   string  true  "Post ID"
// @Param        page  query  int     false "Page number for replies" default(1)
// @Success      200   {object} map[string]interface{}  "posts, master_post, total_pages"
// @Failure      404   {object} map[string]interface{}  "Post not found"
// @Failure      500   {object} map[string]interface{}  "Internal server error"
// @Router       /posts/{id} [get]
func GetPost(c *gin.Context) {
	var post model.Post
	var posts []model.Post
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	if err := database.Database.
		Preload("ReplyTo").
		Preload("ReplyTo.User").
		Preload("Category").
		Preload("User").
		First(&post, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	offset := (page - 1) * pageSize

	if err := database.Database.
		Preload("ReplyTo").
		Preload("ReplyTo.User").
		Preload("Category").
		Preload("User").
		Where("parent_post_id = ? OR id = ?", post.ID, post.ID).
		Order("created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load reactions for each post
	for i := range posts {
		if reactions, userReactions, err := getPostReactions(posts[i].ID, c); err == nil {
			posts[i].Reactions = reactions
			posts[i].UserReactions = userReactions
		}
	}

	var replyCount int64
	database.Database.Model(&model.Post{}).
		Where("parent_post_id = ?", post.ID).
		Count(&replyCount)

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"master_post": post,
		"total_pages": (int(replyCount) + pageSize) / pageSize,
	})
}

// UpdatePost godoc
// @Summary      Update an existing post
// @Description  Fully update an existing post by its ID. Respects master/reply post validation rules.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id      path   int      true  "Post ID"
// @Param        payload body   Payload  true  "Post payload"
// @Success      200     {object} model.Post
// @Failure      400     {object} map[string]interface{}  "Bad request"
// @Failure      403     {object} map[string]interface{}  "Forbidden - user not authorized to update"
// @Failure      404     {object} map[string]interface{}  "Post or category not found"
// @Failure      500     {object} map[string]interface{}  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	var post model.Post
	var payload Payload

	if err := database.Database.First(&post, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDUint := uint(userID.(float64))
	role := c.GetString("role")

	if post.UserID != userIDUint && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this post"})
		return
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.First(&model.Category{}, payload.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if ok, message := validatePost(payload); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	if post.ReplyToID != payload.ReplyToID && payload.ReplyToID != nil {
		var replyToPost model.Post
		if err := database.Database.First(&replyToPost, payload.ReplyToID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reply to post not found"})
			return
		}

		err := services.CreateNotification(replyToPost.UserID, userIDUint, post.ID, "reply")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	post.Title = payload.Title
	post.Content = payload.Content
	post.IsMasterPost = payload.IsMasterPost
	post.ParentPostID = payload.ParentPostID
	post.ReplyToID = payload.ReplyToID
	post.CategoryID = payload.CategoryID

	if err := database.Database.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// PatchPost godoc
// @Summary      Partially update an existing post
// @Description  Updates only the fields provided in the request body. Must pass post ID via the path. Respects user ownership or admin rights.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        id      path   int                       true  "Post ID"
// @Param        payload body   map[string]interface{}    true  "Fields to update"
// @Success      200     {object} model.Post
// @Failure      400     {object} map[string]interface{}  "Bad request"
// @Failure      403     {object} map[string]interface{}  "Forbidden - user not authorized to update"
// @Failure      404     {object} map[string]interface{}  "Post not found"
// @Failure      500     {object} map[string]interface{}  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /posts/{id} [patch]
func PatchPost(c *gin.Context) {
	var post model.Post
	var payload map[string]interface{}

	if err := database.Database.First(&post, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	userID, _ := c.Get("user_id")
	role := c.GetString("role")

	if post.UserID != uint(userID.(float64)) && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this post"})
		return
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if title, ok := payload["title"].(string); ok {
		post.Title = &title
	}

	if content, ok := payload["content"].(string); ok {
		post.Content = content
	}

	if isMasterPost, ok := payload["is_master_post"].(bool); ok {
		post.IsMasterPost = isMasterPost
	}

	if parentPostID, ok := payload["parent_post_id"].(uint); ok {
		post.ParentPostID = &parentPostID
	}

	if replyToID, ok := payload["reply_to_id"].(uint); ok {
		post.ReplyToID = &replyToID
	}

	if categoryID, ok := payload["category_id"].(uint); ok {
		post.CategoryID = categoryID
	}

	// Validate the updated post
	if ok, message := validatePost(post); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	if err := database.Database.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

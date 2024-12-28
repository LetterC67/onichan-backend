package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchPostTitle godoc
// @Summary      Search posts by title
// @Description  Searches for master posts by title within a specified category. Must include `category` and `title` query parameters.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        title      query     string  true  "Title or partial title to search for"
// @Param        category   query     string  true  "Name of the category"
// @Param        page       query     int     false "Page number for pagination" default(1)
// @Success      200  {object}  map[string]interface{}  "{"posts": [...], "total_pages": X}"
// @Failure      400  {object}  map[string]interface{}  "{"error": "Missing required query parameters"}"
// @Failure      404  {object}  map[string]interface{}  "{"error": "Category not found" or "Post not found"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to get post count"}"
// @Router       /search/title [get]
func SearchPostTitle(c *gin.Context) {
	title := c.Query("title")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	categoryName := c.Query("category")

	var category model.Category
	if categoryName != "" {
		if err := database.Database.Where("name = ?", categoryName).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title is required"})
		return
	}

	var posts []model.Post
	offset := (page - 1) * pageSize

	if err := database.Database.
		Preload("User").
		Where("title LIKE ?", "%"+title+"%").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var count int64
	if err := database.Database.
		Model(&model.Post{}).
		Where("title LIKE ?", "%"+title+"%").
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"total_pages": (int(count) + pageSize - 1) / pageSize,
	})
}

// SearchPostReplies godoc
// @Summary      Search replies to a specific post
// @Description  Searches for replies (posts) by content for a given parent post ID. Must provide the parent post's ID via query parameter `id`.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        content  query     string  true  "Content or partial content to search for"
// @Param        id       query     int     true  "Parent post ID"
// @Param        page     query     int     false "Page number for pagination" default(1)
// @Success      200   {object}    map[string]interface{}  "{"posts": [...], "total_pages": X}"
// @Failure      400   {object}    map[string]interface{}  "{"error": "Missing query parameters"}"
// @Failure      404   {object}    map[string]interface{}  "{"error": "Parent post not found" or "Post not found"}"
// @Failure      500   {object}    map[string]interface{}  "{"error": "Failed to get post count"}"
// @Router       /search/posts [get]
func SearchPostReplies(c *gin.Context) {
	content := c.Query("content")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	parentPostID, _ := strconv.Atoi(c.Query("id"))

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content query parameter is required"})
		return
	}

	var parentPost model.Post
	if err := database.Database.First(&parentPost, parentPostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent post not found"})
		return
	}

	var posts []model.Post
	offset := (page - 1) * pageSize

	if err := database.Database.
		Preload("Category").
		Preload("User").
		Where("content LIKE ? AND parent_post_id = ?", "%"+content+"%", parentPostID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	for index := range posts {
		var count int64
		if err := database.Database.
			Model(&model.Post{}).
			Where("parent_post_id = ? AND created_at < ?", parentPostID, posts[index].CreatedAt).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post order"})
			return
		}

		posts[index].Page = (int(count) + pageSize + 1) / pageSize
	}

	var totalCount int64
	if err := database.Database.
		Model(&model.Post{}).
		Where("content LIKE ? AND parent_post_id = ?", "%"+content+"%", parentPostID).
		Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"total_pages": (int(totalCount) + pageSize - 1) / pageSize,
	})
}

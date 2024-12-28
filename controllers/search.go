package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	if err := database.Database.Preload("User").Where("title LIKE ?", "%"+title+"%").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var count int64
	if err := database.Database.Model(&model.Post{}).Where("title LIKE ?", "%"+title+"%").Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"total_pages": (int(count) + pageSize - 1) / pageSize,
	})
}

func SearchPostReplies(c *gin.Context) {
	content := c.Query("content")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	parentPostID, _ := strconv.Atoi(c.Query("id"))

	var parentPost model.Post
	if err := database.Database.First(&parentPost, parentPostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent post not found"})
		return
	}

	var posts []model.Post
	offset := (page - 1) * pageSize

	if err := database.Database.Preload("Category").Preload("User").Where("content LIKE ? AND parent_post_id = ?", "%"+content+"%", parentPostID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	for index := range posts {
		var count int64
		if err := database.Database.Model(&model.Post{}).Where("parent_post_id = ? AND created_at < ?", parentPostID, posts[index].CreatedAt).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post order"})
			return
		}

		posts[index].Page = (int(count) + pageSize + 1) / pageSize
	}

	var count int64
	if err := database.Database.Model(&model.Post{}).Where("content LIKE ? AND parent_post_id = ?", "%"+content+"%", parentPostID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"total_pages": (int(count) + pageSize - 1) / pageSize,
	})
}

package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

func GetUnreadNotifications(c *gin.Context) {
	var notifications []model.Notification

	userID := uint(c.MustGet("user_id").(float64))

	if err := database.Database.Preload("Post").Preload("FromUser").Preload("Post.Category").Where("is_read = false AND user_id = ?", userID).Not("from_user_id", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	for index := range notifications {
		var count int64
		if err := database.Database.Model(&model.Post{}).
			Where("parent_post_id = ? AND created_at < ?", notifications[index].Post.ParentPostID, notifications[index].Post.CreatedAt).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post order"})
			return
		}

		notifications[index].Post.Page = (int(count) + pageSize + 1) / pageSize
	}

	c.JSON(http.StatusOK, notifications)
}

func ReadNotifications(c *gin.Context) {
	userID := uint(c.MustGet("user_id").(float64))

	if err := database.Database.Model(&model.Notification{}).Where("user_id = ? AND is_read = false", userID).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read"})
}

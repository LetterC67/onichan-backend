package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

// GetUnreadNotifications godoc
// @Summary      Get unread notifications for the current user
// @Description  Retrieves all unread notifications for the logged-in user, including preloaded Post, FromUser, and Category information. Automatically calculates and sets the `Page` field for pagination based on the post's creation date.
// @Tags         notifications
// @Produce      json
// @Success      200  {array}   model.Notification
// @Failure      404  {object}  map[string]interface{}  "{"error": "Notification not found"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to get post order"}"
// @Security     ApiKeyAuth
// @Router       /notifications [get]
func GetUnreadNotifications(c *gin.Context) {
	var notifications []model.Notification

	userID := uint(c.MustGet("user_id").(float64))

	if err := database.Database.
		Preload("Post").
		Preload("FromUser").
		Preload("Post.Category").
		Where("is_read = false AND user_id = ?", userID).
		Not("from_user_id", userID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	for i := range notifications {
		var count int64
		if err := database.Database.Model(&model.Post{}).
			Where("parent_post_id = ? AND created_at < ?", notifications[i].Post.ParentPostID, notifications[i].Post.CreatedAt).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post order"})
			return
		}

		notifications[i].Post.Page = (int(count) + pageSize + 1) / pageSize
	}

	c.JSON(http.StatusOK, notifications)
}

// ReadNotifications godoc
// @Summary      Mark all unread notifications as read
// @Description  Marks all unread notifications for the logged-in user as read in the database
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "{"message": "Notifications marked as read"}"
// @Failure      500  {object}  map[string]interface{}  "{"error": "Failed to mark notifications as read"}"
// @Security     ApiKeyAuth
// @Router       /notifications [patch]
func ReadNotifications(c *gin.Context) {
	userID := uint(c.MustGet("user_id").(float64))

	if err := database.Database.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read"})
}

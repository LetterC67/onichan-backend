package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ToggleReaction(c *gin.Context) {
	var payload struct {
		PostID     uint `json:"post_id" binding:"required"`
		ReactionID uint `json:"reaction_id" binding:"required"`
	}

	userID, _ := c.Get("user_id")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reaction model.PostReaction
	err := database.Database.Where("post_id = ? AND user_id = ? AND reaction_id = ?", payload.PostID, userID, payload.ReactionID).First(&reaction).Error

	if err == nil {
		if err := database.Database.Delete(&reaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query reaction"})
		return
	}

	newReaction := model.PostReaction{
		PostID:     payload.PostID,
		UserID:     uint(userID.(float64)),
		ReactionID: payload.ReactionID,
	}

	if err := database.Database.Create(&newReaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction added"})
}

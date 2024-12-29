package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ToggleReactionRequest struct {
	PostID     uint `json:"post_id" binding:"required"`
	ReactionID uint `json:"reaction_id" binding:"required"`
}

// ToggleReaction godoc
// @Summary      Toggle a reaction on a post
// @Description  Adds or removes a user's reaction on a post, depending on whether the user has already reacted.
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        input  body      ToggleReactionRequest  true  "Toggle Reaction Request"
// @Success      200    {object}  map[string]interface{}  "{"message": "Reaction added"} or {"message": "Reaction removed"}"
// @Failure      400    {object}  map[string]interface{}  "{"error": "Bad request"}"
// @Failure      401    {object}  map[string]interface{}  "{"error": "Unauthorized"}"
// @Failure      500    {object}  map[string]interface{}  "{"error": "Failed to add reaction" or "Failed to remove reaction"}"
// @Security     ApiKeyAuth
// @Router       /posts/reactions [put]
func ToggleReaction(c *gin.Context) {
	var payload ToggleReactionRequest

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
	err := database.Database.
		Where("post_id = ? AND user_id = ? AND reaction_id = ?", payload.PostID, userID, payload.ReactionID).
		First(&reaction).Error

	if err == nil {
		if err := database.Database.Unscoped().Delete(&reaction).Error; err != nil {
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

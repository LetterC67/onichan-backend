package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

func ListReactions(c *gin.Context) {
	var reactions []model.Reaction

	if err := database.Database.Find(&reactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, reactions)
}

func GetReaction(c *gin.Context) {
	var reaction model.Reaction
	if err := database.Database.First(&reaction, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
		return
	}

	c.JSON(http.StatusOK, reaction)
}

func CreateReaction(c *gin.Context) {
	var payload struct {
		Name  string `json:"name" binding:"required"`
		Emoji string `json:"emoji" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reaction := model.Reaction{
		Name:  payload.Name,
		Emoji: payload.Emoji,
	}

	if result := database.Database.Create(&reaction); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, "Reaction created successfully")
}

func UpdateReaction(c *gin.Context) {
	var reaction model.Reaction

	if err := database.Database.Find(&reaction, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
		return
	}

	if err := c.ShouldBindJSON(&reaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.Save(&reaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reaction"})
		return
	}

	c.JSON(http.StatusOK, reaction)
}

func PatchReaction(c *gin.Context) {
	var reaction model.Reaction

	if err := database.Database.Find(&reaction, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.Model(&reaction).Updates(&payload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reaction"})
		return
	}

	c.JSON(http.StatusOK, reaction)
}

func DeleteReaction(c *gin.Context) {
	if err := database.Database.Delete(&model.Reaction{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reaction"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

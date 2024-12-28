package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	var user model.User

	if err := database.Database.Find(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetAllAvatars(c *gin.Context) {
	var avatars []model.Avatar

	if err := database.Database.Find(&avatars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve avatars"})
		return
	}

	c.JSON(http.StatusOK, avatars)
}

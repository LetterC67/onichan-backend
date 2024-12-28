package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieves a single user by their ID
// @Tags         users
// @Produce      json
// @Param        id   path      int   true  "User ID"
// @Success      200  {object}  model.User
// @Failure      404  {object}  map[string]interface{}  "{"error":"User not found"}"
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
	var user model.User

	if err := database.Database.Find(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllAvatars godoc
// @Summary      List all available avatars
// @Description  Retrieves an array of all avatars in the system
// @Tags         users
// @Produce      json
// @Success      200  {array}   model.Avatar
// @Failure      500  {object}  map[string]interface{}  "{"error":"Failed to retrieve avatars"}"
// @Router       /users/avatars [get]
func GetAllAvatars(c *gin.Context) {
	var avatars []model.Avatar

	if err := database.Database.Find(&avatars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve avatars"})
		return
	}

	c.JSON(http.StatusOK, avatars)
}

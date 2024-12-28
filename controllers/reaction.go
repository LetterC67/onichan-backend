package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"

	"github.com/gin-gonic/gin"
)

// ListReactions godoc
// @Summary      List all reactions
// @Description  Retrieves a list of all reaction objects in the database
// @Tags         reactions
// @Produce      json
// @Success      200  {array}   model.Reaction
// @Failure      500  {object}  map[string]interface{}  "Failed to retrieve categories"
// @Router       /reactions [get]
func ListReactions(c *gin.Context) {
	var reactions []model.Reaction

	if err := database.Database.Find(&reactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, reactions)
}

// GetReaction godoc
// @Summary      Get a reaction by ID
// @Description  Retrieves a single reaction by its ID
// @Tags         reactions
// @Produce      json
// @Param        id   path      int   true  "Reaction ID"
// @Success      200  {object}  model.Reaction
// @Failure      404  {object}  map[string]interface{}  "Reaction not found"
// @Router       /reactions/{id} [get]
func GetReaction(c *gin.Context) {
	var reaction model.Reaction
	if err := database.Database.First(&reaction, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
		return
	}

	c.JSON(http.StatusOK, reaction)
}

type CreateReactionRequest struct {
	Name  string `json:"name" binding:"required"`
	Emoji string `json:"emoji" binding:"required"`
}

// CreateReaction godoc
// @Summary      Create a new reaction
// @Description  Creates a new reaction with the specified name and emoji
// @Tags         reactions
// @Accept       json
// @Produce      json
// @Param        payload  body      CreateReactionRequest  true  "Create Reaction Request"
// @Success      200      {string}  string                 "Reaction created successfully"
// @Failure      400      {object}  map[string]interface{} "Bad request"
// @Failure      500      {object}  map[string]interface{} "Failed to create reaction"
// @Security     ApiKeyAuth
// @Router       /reactions [post]
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

// UpdateReaction godoc
// @Summary      Update an existing reaction (full update)
// @Description  Updates all fields of a reaction by its ID
// @Tags         reactions
// @Accept       json
// @Produce      json
// @Param        id   path      int   true  "Reaction ID"
// @Param        data body      model.Reaction true  "Reaction update payload"
// @Success      200  {object}  model.Reaction
// @Failure      400  {object}  map[string]interface{} "Bad request"
// @Failure      404  {object}  map[string]interface{} "Reaction not found"
// @Failure      500  {object}  map[string]interface{} "Failed to update reaction"
// @Security     ApiKeyAuth
// @Router       /reactions/{id} [put]
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

// PatchReaction godoc
// @Summary      Partially update an existing reaction
// @Description  Updates only the fields provided in the request body
// @Tags         reactions
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "Reaction ID"
// @Param        data  body      map[string]interface{} true  "Partial reaction update payload"
// @Success      200   {object}  model.Reaction
// @Failure      400   {object}  map[string]interface{}  "Bad request"
// @Failure      404   {object}  map[string]interface{}  "Reaction not found"
// @Failure      500   {object}  map[string]interface{}  "Failed to update reaction"
// @Security     ApiKeyAuth
// @Router       /reactions/{id} [patch]
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

	if err := database.Database.Model(&reaction).Updates(payload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reaction"})
		return
	}

	c.JSON(http.StatusOK, reaction)
}

// DeleteReaction godoc
// @Summary      Delete a reaction by ID
// @Description  Removes a reaction from the database by its ID
// @Tags         reactions
// @Produce      json
// @Param        id   path      int  true  "Reaction ID"
// @Success      204  "No Content"
// @Failure      500  {object}  map[string]interface{}  "Failed to delete reaction"
// @Security     ApiKeyAuth
// @Router       /reactions/{id} [delete]
func DeleteReaction(c *gin.Context) {
	if err := database.Database.Delete(&model.Reaction{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reaction"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

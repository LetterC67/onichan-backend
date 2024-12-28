package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateCategoryRequest represents the payload required to create a category.
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	ImageURL    string `json:"image_url"`
}

// ListCategories godoc
// @Summary      List all categories
// @Description  Returns an array of all categories
// @Tags         categories
// @Produce      json
// @Success      200  {array}   model.Category
// @Failure      500  {object}  map[string]interface{}  "{"error":"Failed to retrieve categories"}"
// @Router       /categories [get]
func ListCategories(c *gin.Context) {
	var categories []model.Category

	if err := database.Database.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func isInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// GetCategory godoc
// @Summary      Get a specific category by ID or name
// @Description  If `id` is an integer, it searches by ID. Otherwise, it searches by name.
// @Tags         categories
// @Produce      json
// @Param        id   path      string  true  "Category ID or name"
// @Success      200  {object}  model.Category
// @Failure      404  {object}  map[string]interface{}  "{"error":"Category not found"}"
// @Router       /categories/{id} [get]
func GetCategory(c *gin.Context) {
	var category model.Category
	query := c.Param("id")

	if isInteger(query) {
		// Search by ID
		if err := database.Database.First(&category, query).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	} else {
		// Search by name
		if err := database.Database.Where("name = ?", query).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Creates a new category with the provided name, description, and optional image URL
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      CreateCategoryRequest  true  "Category creation payload"
// @Success      200       {string}  string                "Category created successfully"
// @Failure      400       {object}  map[string]interface{} "{"error":"Bad request"}"
// @Failure      500       {object}  map[string]interface{} "{"error":"Internal Server Error"}"
// @Security     ApiKeyAuth
// @Router       /categories [post]
func CreateCategory(c *gin.Context) {
	var payload CreateCategoryRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := model.Category{
		Name:        payload.Name,
		Description: payload.Description,
		ImageURL:    &payload.ImageURL,
	}

	if result := database.Database.Create(&category); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, "Category created successfully")
}

// UpdateCategoryRequest represents the full update payload for a category.
type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	ImageURL    string `json:"image_url"`
}

// UpdateCategory godoc
// @Summary      Update an existing category (full update)
// @Description  Updates all fields of a category by ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "Category ID"
// @Param        data  body      UpdateCategoryRequest  true  "Category update payload"
// @Success      200   {object}  model.Category
// @Failure      400   {object}  map[string]interface{}  "{"error":"Bad request"}"
// @Failure      404   {object}  map[string]interface{}  "{"error":"Category not found"}"
// @Failure      500   {object}  map[string]interface{}  "{"error":"Failed to update category"}"
// @Security     ApiKeyAuth
// @Router       /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	var category model.Category

	if err := database.Database.First(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var payload UpdateCategoryRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = payload.Name
	category.Description = payload.Description
	category.ImageURL = &payload.ImageURL

	if err := database.Database.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// PatchCategory godoc
// @Summary      Partially update a category
// @Description  Updates only the provided fields in the request body
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id    path      int                true  "Category ID"
// @Param        data  body      map[string]interface{}  true  "Partial category update payload"
// @Success      200   {object}  model.Category
// @Failure      400   {object}  map[string]interface{}  "{"error":"Bad request"}"
// @Failure      404   {object}  map[string]interface{}  "{"error":"Category not found"}"
// @Failure      500   {object}  map[string]interface{}  "{"error":"Failed to update category"}"
// @Security     ApiKeyAuth
// @Router       /categories/{id} [patch]
func PatchCategory(c *gin.Context) {
	var category model.Category

	if err := database.Database.Find(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.Model(&category).Updates(payload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary      Delete a category by ID
// @Description  Removes a category from the database by ID
// @Tags         categories
// @Produce      json
// @Param        id  path      int  true  "Category ID"
// @Success      204  "No Content"
// @Failure      500  {object}  map[string]interface{}  "{"error":"Failed to delete category"}"
// @Security     ApiKeyAuth
// @Router       /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	if err := database.Database.Delete(&model.Category{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

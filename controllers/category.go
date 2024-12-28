package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

func GetCategory(c *gin.Context) {
	var category model.Category
	query := c.Param("id")

	if isInteger(query) {
		if err := database.Database.First(&category, query).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	} else {
		if err := database.Database.Where("name = ?", query).First(&category).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
	}

	c.JSON(http.StatusOK, category)
}

func CreateCategory(c *gin.Context) {
	var payload struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description" binding:"required"`
		ImageURL    string `json:"image_url"`
	}

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

func UpdateCategory(c *gin.Context) {
	var category model.Category

	if err := database.Database.First(&category, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Database.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

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

	if err := database.Database.Model(&category).Updates(&payload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	if err := database.Database.Delete(&model.Category{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

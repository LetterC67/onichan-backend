package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadImage godoc
// @Summary      Upload an image
// @Description  Accepts a single image file (jpg, jpeg, png, gif, webp, svg, bmp) via multipart/form-data.
// @Tags         upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "File to upload"
// @Success      200   {object}  map[string]interface{}  "{"message": "File uploaded successfully", "path": "uploads/<filename>"}"
// @Failure      400   {object}  map[string]interface{}  "{"error": "No file uploaded or invalid file format"}"
// @Failure      500   {object}  map[string]interface{}  "{"error": "Failed to save file"}"
// @Router       /upload [post]
func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" && ext != ".svg" && ext != ".bmp" {
		c.JSON(400, gin.H{"error": "Invalid file format"})
		return
	}

	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	dst := filepath.Join(uploadPath, uniqueFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"path":    dst,
	})
}

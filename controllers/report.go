package controllers

import (
	"net/http"
	"onichan/database"
	"onichan/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateReportRequest struct {
	PostID uint `json:"post_id" binding:"required"`
}

// CreateReport godoc
// @Summary      Create a new report for a post
// @Description  Creates a report for a specific post by ID
// @Tags         reports
// @Accept       json
// @Produce      json
// @Param        payload body      CreateReportRequest true "Report creation payload"
// @Success      200     {object}  map[string]interface{}  "{"message": "Report created successfully"}"
// @Failure      400     {object}  map[string]interface{}  "{"error": "Bad request"}"
// @Failure      404     {object}  map[string]interface{}  "{"error": "Post not found"}"
// @Failure      500     {object}  map[string]interface{}  "{"error": "Failed to create report"}"
// @Security     ApiKeyAuth
// @Router       /reports [post]
func CreateReport(c *gin.Context) {
	var payload struct {
		PostID uint `json:"post_id" binding:"required"`
	}
	userID := uint(c.MustGet("user_id").(float64))

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var post model.Post
	if err := database.Database.First(&post, payload.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if post.IsDeleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post has been removed already"})
		return
	}

	report := model.Report{
		PostID: payload.PostID,
		UserID: userID,
	}

	if result := database.Database.Create(&report); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report created successfully",
	})
}

// ListReports godoc
// @Summary      List all reports
// @Description  Returns a paginated list of reports, including the reported Post and the reporting User
// @Tags         reports
// @Produce      json
// @Param        page  query     int  false  "Page number" default(1)
// @Success      200   {object}  map[string]interface{}  "{"reports": [...], "total_pages": X}"
// @Failure      500   {object}  map[string]interface{}  "{"error": "Failed to retrieve reports"}"
// @Security     ApiKeyAuth
// @Router       /reports [get]
func ListReports(c *gin.Context) {
	var reports []model.Report
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	offset := (page - 1) * pageSize

	if err := database.Database.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Preload("User").
		Preload("Post").
		Preload("Post.User").
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reports"})
		return
	}

	var count int64
	if err := database.Database.Model(&model.Report{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get report count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":     reports,
		"total_pages": (int(count) + pageSize - 1) / pageSize,
	})
}

type ResolveReportRequest struct {
	ReportID   uint `json:"report_id" binding:"required"`
	DeletePost bool `json:"delete_post"`
}

// ResolveReport godoc
// @Summary      Resolve a report
// @Description  Resolves a report by marking it as resolved. Optionally deletes the associated post if specified.
// @Tags         reports
// @Accept       json
// @Produce      json
// @Param        payload  body      ResolveReportRequest  true  "Resolve Report Request"
// @Success      200      {object}  map[string]interface{}  "{"message": "Report resolved successfully"}"
// @Failure      400      {object}  map[string]interface{}  "{"error": "Bad request"}"
// @Failure      404      {object}  map[string]interface{}  "{"error": "Report or post not found"}"
// @Failure      500      {object}  map[string]interface{}  "{"error": "Failed to resolve report"}"
// @Security     ApiKeyAuth
// @Router       /reports [patch]
func ResolveReport(c *gin.Context) {
	var payload struct {
		ReportID   uint `json:"report_id" binding:"required"`
		DeletePost bool `json:"delete_post"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var report model.Report
	var post model.Post

	if err := database.Database.First(&report, payload.ReportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	if payload.DeletePost {
		if err := database.Database.First(&post, report.PostID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		post.Content = "[This post has been deleted by a moderator]"
		post.IsDeleted = true

		if post.Title != nil {
			*post.Title = "[This post has been deleted by a moderator]"
		}

		if result := database.Database.Save(&post); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	report.Resolved = true
	if result := database.Database.Save(&report); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report resolved successfully"})
}

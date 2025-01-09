package main

import (
	"onichan/controllers"
	"onichan/database"
	_ "onichan/docs"
	"onichan/middleware"
	"onichan/services"
	"onichan/utils"
	"onichan/websocket"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	utils.LoadEnv()
	utils.LoadJWT()
	services.LoadEnv()
	database.Connect()
	controllers.LoadPageSize()

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	maxFileSize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	r.MaxMultipartMemory = int64(maxFileSize)

	r.Use(middleware.CORSMiddleware())

	api := r.Group("api")

	api.POST("/upload", middleware.JWTMiddleware(database.Database), controllers.UploadImage)
	api.Static("/uploads", os.Getenv("UPLOAD_PATH"))

	authRoute := api.Group("/auth")
	{
		authRoute.POST("/register", middleware.CORSMiddleware(), controllers.Register)
		authRoute.POST("/login", middleware.CORSMiddleware(), controllers.Login)
		authRoute.PATCH("/change-password", middleware.JWTMiddleware(database.Database), controllers.ChangePassword)
		authRoute.PATCH("/change-email", middleware.JWTMiddleware(database.Database), controllers.ChangeEmail)
		authRoute.PATCH("/change-avatar", middleware.JWTMiddleware(database.Database), controllers.ChangeAvatar)
		authRoute.POST("/forgot-password", controllers.ForgotPassword)
		authRoute.POST("/reset-password", controllers.ResetForgottenPassword)
	}

	userRoute := api.Group("/users")
	{
		userRoute.GET("/:id", controllers.GetUser)
		userRoute.GET("/avatars", controllers.GetAllAvatars)
	}

	categoryRoute := api.Group("/categories")
	{
		categoryRoute.POST("", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.CreateCategory)
		categoryRoute.GET("", controllers.ListCategories)
		categoryRoute.GET("/:id", controllers.GetCategory)
		categoryRoute.PUT("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.UpdateCategory)
		categoryRoute.PATCH("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.PatchCategory)
		categoryRoute.DELETE("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.DeleteCategory)
	}

	reactionRoute := api.Group("/reactions")
	{
		reactionRoute.POST("", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.CreateReaction)
		reactionRoute.GET("", controllers.ListReactions)
		reactionRoute.GET("/:id", controllers.GetReaction)
		reactionRoute.PUT("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.UpdateReaction)
		reactionRoute.PATCH("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.PatchReaction)
		reactionRoute.DELETE("/:id", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.DeleteReaction)
	}

	postRoute := api.Group("/posts")
	{
		postRoute.POST("", middleware.JWTMiddleware(database.Database), controllers.CreatePost)
		postRoute.GET("", controllers.ListPosts)
		postRoute.GET("/:id", controllers.GetPost)
		postRoute.PUT("/:id", middleware.JWTMiddleware(database.Database), controllers.UpdatePost)
		postRoute.PATCH("/:id", middleware.JWTMiddleware(database.Database), controllers.PatchPost)
		postRoute.PUT("/reactions", middleware.JWTMiddleware(database.Database), controllers.ToggleReaction)
	}

	notificationRoute := api.Group("notifications")
	{
		notificationRoute.GET("", middleware.JWTMiddleware(database.Database), controllers.GetUnreadNotifications)
		notificationRoute.PATCH("", middleware.JWTMiddleware(database.Database), controllers.ReadNotifications)
	}

	searchRoute := api.Group("/search")
	{
		searchRoute.GET("/title", controllers.SearchPostTitle)
		searchRoute.GET("/posts", controllers.SearchPostReplies)
	}

	reportRoute := api.Group("/reports")
	{
		reportRoute.POST("", middleware.JWTMiddleware(database.Database), controllers.CreateReport)
		reportRoute.GET("", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.ListReports)
		reportRoute.PATCH("", middleware.JWTMiddleware(database.Database), middleware.AdminOnly(), controllers.ResolveReport)
	}

	r.GET("/ws", websocket.WsHandler)

	r.Run(":" + os.Getenv("APP_PORT"))
}

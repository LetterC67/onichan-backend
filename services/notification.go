package services

import (
	"onichan/database"
	"onichan/model"
	"onichan/utils"
	"onichan/websocket"
)

func CreateNotification(forUser, fromUser uint, postID uint, notificationType string) error {
	notification := model.Notification{
		UserID:           forUser,
		FromUserID:       fromUser,
		PostID:           postID,
		NotificationType: notificationType,
	}

	if err := database.Database.Create(&notification).Error; err != nil {
		return err
	}

	database.Database.Model(&notification).Preload("FromUser").Preload("Post").Preload("Post.Category").First(&notification)
	notification.Post.Page = utils.GetPostPage(notification.Post)

	websocket.SendWebSocketNotification(forUser, notification)

	return nil
}

package model

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	UserID           uint      `gorm:"index:user_notifications_index,priority:1"`
	FromUserID       uint      `json:"from_user_id"`
	PostID           uint      `json:"post_id"`
	FromUser         User      `gorm:"foreignKey:FromUserID;references:ID" json:"from_user"`
	Post             Post      `gorm:"foreignKey:PostID;references:ID" json:"post"`
	NotificationType string    `json:"notification_type"`
	CreatedAt        time.Time `gorm:"index:user_notifications_index,priority:2" json:"created_at"`
	IsRead           bool      `gorm:"default:false;index:user_notifications_index,priority:3" json:"is_read"`
}

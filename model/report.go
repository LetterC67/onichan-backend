package model

import "gorm.io/gorm"

type Report struct {
	gorm.Model

	PostID   uint `json:"post_id"`
	Post     Post `gorm:"foreignKey:PostID" json:"post"`
	UserID   uint `json:"user_id"`
	User     User `gorm:"foreignKey:UserID" json:"user"`
	Resolved bool `gorm:"default:false" json:"resolved"`
}

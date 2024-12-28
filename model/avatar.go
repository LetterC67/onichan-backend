package model

import "gorm.io/gorm"

type Avatar struct {
	gorm.Model
	AvatarURL string `gorm:"size:255;unique" json:"avatar_url"`
}

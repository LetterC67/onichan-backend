package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string  `gorm:"size:31;not null;uniqueIndex" json:"username"`
	Email        string  `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash string  `gorm:"size:255;not null" json:"-"`
	Salt         string  `gorm:"size:255;not null" json:"-"`
	AvatarURL    *string `gorm:"size:255" json:"avatar_url"`
	Bio          *string `gorm:"type:text" json:"bio"`
	Role         string  `gorm:"size:31;default:'user'" json:"role"`
}

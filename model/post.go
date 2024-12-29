package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name        string  `gorm:"size:63;not null;unique" json:"name"`
	Description string  `gorm:"size:255;not null" json:"description"`
	ImageURL    *string `gorm:"size:255" json:"image_url"`
	Posts       []Post  `gorm:"foreignKey:CategoryID"`
}

type Reaction struct {
	gorm.Model
	Name  string `gorm:"size:31;not null;unique" json:"name"`
	Emoji string `gorm:"type:varchar(15)" json:"emoji"`
}

type PostReaction struct {
	gorm.Model
	PostID     uint     `gorm:"index:post_user_reaction_index;constraint:OnDelete:CASCADE" json:"post_id"`
	UserID     uint     `gorm:"index:post_user_reaction_index" json:"user_id"`
	ReactionID uint     `gorm:"index:post_user_reaction_index" json:"reaction_id"`
	Reaction   Reaction `gorm:"foreignKey:ReactionID" json:"reaction"`
}

type PostReactionCount struct {
	Reaction Reaction `gorm:"-" json:"reaction"`
	Count    int      `gorm:"-" json:"count"`
}

type Post struct {
	gorm.Model
	UserID        uint                `json:"user_id"`
	User          User                `gorm:"foreignKey:UserID" json:"user"`
	Title         *string             `gorm:"type:text;index" json:"title"`
	Content       string              `gorm:"type:text" json:"content"`
	IsMasterPost  bool                `gorm:"index" json:"is_master_post"`
	ParentPostID  *uint               `gorm:"index;onDelete:SET NULL" json:"parent_post_id"`
	ParentPost    *Post               `gorm:"foreignKey:ParentPostID" json:"-"`
	ChildrenPosts []Post              `gorm:"foreignKey:ParentPostID" json:"-"`
	ReplyToID     *uint               `json:"reply_to_id"`
	ReplyTo       *Post               `gorm:"foreignKey:ReplyToID" json:"reply_to"`
	CreatedAt     time.Time           `gorm:"index:post_category_index,priority:1;" json:"created_at"`
	LastUpdated   time.Time           `gorm:"index" json:"last_updated"`
	CategoryID    uint                `gorm:"index:post_category_index,priority:2;" json:"category_id"`
	Category      Category            `gorm:"foreignKey:CategoryID" json:"category"`
	RepliesCount  int                 `gorm:"-" json:"replies"`
	Reactions     []PostReactionCount `gorm:"-" json:"reactions"`
	UserReactions []PostReactionCount `gorm:"-" json:"user_reactions"`
	Page          int                 `gorm:"-" json:"page"`
	IsDeleted     bool                `gorm:"default:false" json:"is_deleted"`
}

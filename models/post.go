package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text"`
	Summary   string         `json:"summary"`
	Status    string         `json:"status" gorm:"default:'draft'"` // draft, published, archived
	ViewCount int            `json:"view_count" gorm:"default:0"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type PostResponse struct {
	ID        uint         `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Summary   string       `json:"summary"`
	Status    string       `json:"status"`
	ViewCount int          `json:"view_count"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (p *Post) ToResponse() PostResponse {
	return PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Summary:   p.Summary,
		Status:    p.Status,
		ViewCount: p.ViewCount,
		User:      p.User.ToResponse(),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
} 
package models

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" binding:"required"`
	LogoURL     string         `json:"logo_url"`
	FoundedYear int            `json:"founded_year"`
	StadiumAddr string         `json:"stadium_address"`
	City        string         `json:"city"`
	Players     []Player       `json:"players,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Team) TableName() string { return "teams" }

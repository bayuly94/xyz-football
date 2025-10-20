package models

import (
	"time"

	"gorm.io/gorm"
)

type Goal struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MatchID   uint      `json:"match_id" binding:"required"`
	PlayerID  uint      `json:"player_id" binding:"required"`
	Minute    int       `json:"minute" binding:"required,min=0,max=130"`
	CreatedAt time.Time `json:"created_at"`

	Match  Match  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Player Player `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Goal) TableName() string { return "goals" }

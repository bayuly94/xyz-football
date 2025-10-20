package models

import (
	"time"

	"gorm.io/gorm"
)

type PlayerPosition string

const (
	Striker    PlayerPosition = "striker"
	Midfield   PlayerPosition = "midfield"
	Defender   PlayerPosition = "defender"
	Goalkeeper PlayerPosition = "goalkeeper"
)

type Player struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	TeamID    uint           `json:"team_id" binding:"required"`
	Name      string         `json:"name" binding:"required"`
	HeightCM  float64        `json:"height_cm"`
	WeightKG  float64        `json:"weight_kg"`
	Position  PlayerPosition `json:"position" binding:"required,oneof=striker midfielder defender goalkeeper"`
	Number    int            `json:"number" binding:"required,min=1,max=99"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Team Team `json:"team" gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Player) TableName() string { return "players" }

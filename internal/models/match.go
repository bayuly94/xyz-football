package models

import (
	"time"

	"gorm.io/gorm"
)

type MatchStatus string

const (
	Scheduled MatchStatus = "scheduled"
	Finished  MatchStatus = "finished"
)

type Match struct {
	ID         uint        `json:"id" gorm:"primaryKey"`
	MatchTime  time.Time   `json:"match_time" binding:"required"` // tanggal + waktu
	HomeTeamID uint        `json:"home_team_id" binding:"required"`
	AwayTeamID uint        `json:"away_team_id" binding:"required,nefield=HomeTeamID"`
	HomeScore  *int        `json:"home_score,omitempty"` // nullable bila belum selesai
	AwayScore  *int        `json:"away_score,omitempty"`
	Status     MatchStatus `json:"status" gorm:"default:scheduled"`
	Goals      []Goal      `json:"goals,omitempty"`

	HomeTeam Team `json:"home_team" gorm:"foreignKey:HomeTeamID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	AwayTeam Team `json:"away_team" gorm:"foreignKey:AwayTeamID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (Match) TableName() string { return "matches" }

package repositories

import (
	"gorm.io/gorm"
	"xyz-football/internal/models"
)

type GoalRepository interface {
	Create(goal *models.Goal) error
	FindByID(id uint) (*models.Goal, error)
	FindByMatch(matchID uint) ([]models.Goal, error)
	FindByPlayer(playerID uint) ([]models.Goal, error)
	FindByTeam(teamID uint) ([]models.Goal, error)
	Update(goal *models.Goal) error
	Delete(id uint) error
}

type goalRepository struct {
	db *gorm.DB
}

func NewGoalRepository(db *gorm.DB) GoalRepository {
	return &goalRepository{db: db}
}

func (r *goalRepository) Create(goal *models.Goal) error {
	return r.db.Create(goal).Error
}

func (r *goalRepository) FindByID(id uint) (*models.Goal, error) {
	var goal models.Goal
	err := r.db.Preload("Player").Preload("Match").First(&goal, id).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *goalRepository) FindByMatch(matchID uint) ([]models.Goal, error) {
	var goals []models.Goal
	err := r.db.
		Preload("Player").
		Where("match_id = ?", matchID).
		Order("minute ASC").
		Find(&goals).Error
	return goals, err
}

func (r *goalRepository) FindByPlayer(playerID uint) ([]models.Goal, error) {
	var goals []models.Goal
	err := r.db.
		Preload("Match").
		Where("player_id = ?", playerID).
		Find(&goals).Error
	return goals, err
}

func (r *goalRepository) FindByTeam(teamID uint) ([]models.Goal, error) {
	var goals []models.Goal
	err := r.db.
		Joins("JOIN matches ON matches.id = goals.match_id").
		Where("matches.home_team_id = ? OR matches.away_team_id = ?", teamID, teamID).
		Preload("Player").
		Preload("Match").
		Find(&goals).Error
	return goals, err
}

func (r *goalRepository) Update(goal *models.Goal) error {
	return r.db.Save(goal).Error
}

func (r *goalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Goal{}, id).Error
}

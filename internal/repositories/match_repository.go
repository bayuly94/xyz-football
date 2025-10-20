package repositories

import (
	"time"

	"gorm.io/gorm"
	"xyz-football/internal/models"
)

type MatchRepository interface {
	Create(match *models.Match) error
	FindAll() ([]models.Match, error)
	FindByID(id uint) (*models.Match, error)
	FindByDateRange(start, end time.Time) ([]models.Match, error)
	FindByTeamID(teamID uint) ([]models.Match, error)
	Update(match *models.Match) error
	Delete(id uint) error
	WithTransaction(txFunc func(repo MatchRepository) error) error
	GetDB() *gorm.DB
}

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) Create(match *models.Match) error {
	if err := r.db.Create(match).Error; err != nil {
		return err
	}
	// Fetch the created match with all relationships
	return r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Preload("Goals").
		First(match, match.ID).Error
}

func (r *matchRepository) FindAll() ([]models.Match, error) {
	var matches []models.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Preload("Goals").
		Order("match_time ASC").
		Find(&matches).Error
	return matches, err
}

func (r *matchRepository) FindByID(id uint) (*models.Match, error) {
	var match models.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Preload("Goals").
		First(&match, id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) FindByDateRange(start, end time.Time) ([]models.Match, error) {
	var matches []models.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Where("match_time BETWEEN ? AND ?", start, end).
		Order("match_time ASC").
		Find(&matches).Error
	return matches, err
}

func (r *matchRepository) FindByTeamID(teamID uint) ([]models.Match, error) {
	var matches []models.Match
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Where("home_team_id = ? OR away_team_id = ?", teamID, teamID).
		Order("match_time ASC").
		Find(&matches).Error
	return matches, err
}

func (r *matchRepository) Update(match *models.Match) error {
	if err := r.db.Save(match).Error; err != nil {
		return err
	}
	// Fetch the updated match with all relationships
	return r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Preload("Goals").
		First(match, match.ID).Error
}

func (r *matchRepository) Delete(id uint) error {
	return r.db.Delete(&models.Match{}, id).Error
}

func (r *matchRepository) WithTransaction(txFunc func(repo MatchRepository) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txRepo := &matchRepository{db: tx}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := txFunc(txRepo); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *matchRepository) GetDB() *gorm.DB {
	return r.db
}

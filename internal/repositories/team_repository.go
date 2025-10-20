package repositories

import (
	"xyz-football/internal/models"

	"gorm.io/gorm"
)

type TeamRepository interface {
	Create(team *models.Team) error
	FindAll() ([]models.Team, error)
	FindByID(id uint) (*models.Team, error)
	Update(team *models.Team) error
	Delete(id uint) error
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) FindAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Find(&teams).Error
	return teams, err
}

func (r *teamRepository) FindByID(id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) Update(team *models.Team) error {
	// check is exist
	if _, err := r.FindByID(team.ID); err != nil {
		return err
	}
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(id uint) error {
	// First check if team exists
	team, err := r.FindByID(id)
	if err != nil {
		return err
	}
	
	// Use Unscoped() to ensure we can find soft-deleted records if needed
	// and Delete() will perform a soft delete because of gorm.DeletedAt
	return r.db.Delete(team).Error
}

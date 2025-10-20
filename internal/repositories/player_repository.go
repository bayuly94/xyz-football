package repositories

import (
	"gorm.io/gorm"
	"xyz-football/internal/models"
)

type PlayerRepository interface {
	Create(player *models.Player) error
	FindAll() ([]models.Player, error)
	FindByID(id uint) (*models.Player, error)
	FindByTeam(teamID uint) ([]models.Player, error)
	Update(player *models.Player) error
	Delete(id uint) error
}

type playerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

func (r *playerRepository) Create(player *models.Player) error {
	if err := r.db.Create(player).Error; err != nil {
		return err
	}
	// Fetch the created player with team data
	return r.db.Preload("Team").First(player, player.ID).Error
}

func (r *playerRepository) FindAll() ([]models.Player, error) {
	var players []models.Player
	err := r.db.Preload("Team").Find(&players).Error
	return players, err
}

func (r *playerRepository) FindByID(id uint) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("Team").First(&player, id).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) FindByTeam(teamID uint) ([]models.Player, error) {
	var players []models.Player
	err := r.db.Where("team_id = ?", teamID).Find(&players).Error
	return players, err
}

func (r *playerRepository) Update(player *models.Player) error {
	if err := r.db.Save(player).Error; err != nil {
		return err
	}
	// Fetch the updated player with team data
	return r.db.Preload("Team").First(player, player.ID).Error
}

func (r *playerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Player{}, id).Error
}

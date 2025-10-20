package services

import (
	"errors"

	"xyz-football/internal/models"
	"xyz-football/internal/repositories"
)

type PlayerService interface {
	CreatePlayer(player *models.Player) error
	GetAllPlayers() ([]models.Player, error)
	GetPlayerByID(id uint) (*models.Player, error)
	GetPlayersByTeam(teamID uint) ([]models.Player, error)
	UpdatePlayer(player *models.Player) error
	DeletePlayer(id uint) error
}

type playerService struct {
	repo repositories.PlayerRepository
}

func NewPlayerService(repo repositories.PlayerRepository) PlayerService {
	return &playerService{repo: repo}
}

func (s *playerService) CreatePlayer(player *models.Player) error {
	// Validate player number is unique within team
	existingPlayers, err := s.repo.FindByTeam(player.TeamID)
	if err != nil {
		return err
	}

	for _, p := range existingPlayers {
		if p.Number == player.Number {
			return errors.New("player number already exists in this team")
		}
	}

	return s.repo.Create(player)
}

func (s *playerService) GetAllPlayers() ([]models.Player, error) {
	return s.repo.FindAll()
}

func (s *playerService) GetPlayerByID(id uint) (*models.Player, error) {
	return s.repo.FindByID(id)
}

func (s *playerService) GetPlayersByTeam(teamID uint) ([]models.Player, error) {
	return s.repo.FindByTeam(teamID)
}

func (s *playerService) UpdatePlayer(player *models.Player) error {
	// Check if player exists
	_, err := s.repo.FindByID(player.ID)
	if err != nil {
		return errors.New("player not found")
	}

	// Validate player number is unique within team (if number is being updated)
	existingPlayers, err := s.repo.FindByTeam(player.TeamID)
	if err != nil {
		return err
	}

	for _, p := range existingPlayers {
		if p.ID != player.ID && p.Number == player.Number {
			return errors.New("player number already exists in this team")
		}
	}

	return s.repo.Update(player)
}

func (s *playerService) DeletePlayer(id uint) error {
	// Check if player exists
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("player not found")
	}

	return s.repo.Delete(id)
}

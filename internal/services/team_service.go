package services

import (
	"xyz-football/internal/models"
	"xyz-football/internal/repositories"
)

type TeamService interface {
	CreateTeam(team *models.Team) error
	GetAllTeams() ([]models.Team, error)
	GetTeamByID(id uint) (*models.Team, error)
	UpdateTeam(team *models.Team) error
	DeleteTeam(id uint) error
}

type teamService struct {
	repo repositories.TeamRepository
}

func NewTeamService(repo repositories.TeamRepository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) CreateTeam(team *models.Team) error {
	return s.repo.Create(team)
}

func (s *teamService) GetAllTeams() ([]models.Team, error) {
	return s.repo.FindAll()
}

func (s *teamService) GetTeamByID(id uint) (*models.Team, error) {
	return s.repo.FindByID(id)
}

func (s *teamService) UpdateTeam(team *models.Team) error {
	return s.repo.Update(team)
}

func (s *teamService) DeleteTeam(id uint) error {
	return s.repo.Delete(id)
}

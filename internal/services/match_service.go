package services

import (
	"errors"
	"time"

	"xyz-football/internal/models"
	"xyz-football/internal/repositories"
)

type MatchService interface {
	CreateMatch(match *models.Match) error
	GetAllMatches() ([]models.Match, error)
	GetMatchByID(id uint) (*models.Match, error)
	GetMatchesByDateRange(start, end time.Time) ([]models.Match, error)
	GetMatchesByTeam(teamID uint) ([]models.Match, error)
	UpdateMatch(match *models.Match) error
	DeleteMatch(id uint) error
	ReportMatchResult(matchID uint, homeScore, awayScore int, goals []models.Goal) error
}

type matchService struct {
	repo     repositories.MatchRepository
	goalRepo repositories.GoalRepository
}

func NewMatchService(matchRepo repositories.MatchRepository, goalRepo repositories.GoalRepository) MatchService {
	return &matchService{
		repo:     matchRepo,
		goalRepo: goalRepo,
	}
}

func (s *matchService) CreateMatch(match *models.Match) error {
	// Validate teams are different
	if match.HomeTeamID == match.AwayTeamID {
		return errors.New("home and away teams must be different")
	}

	// Set default status if not provided
	if match.Status == "" {
		match.Status = models.Scheduled
	}

	return s.repo.Create(match)
}

func (s *matchService) GetAllMatches() ([]models.Match, error) {
	return s.repo.FindAll()
}

func (s *matchService) GetMatchByID(id uint) (*models.Match, error) {
	return s.repo.FindByID(id)
}

func (s *matchService) GetMatchesByDateRange(start, end time.Time) ([]models.Match, error) {
	return s.repo.FindByDateRange(start, end)
}

func (s *matchService) GetMatchesByTeam(teamID uint) ([]models.Match, error) {
	return s.repo.FindByTeamID(teamID)
}

func (s *matchService) UpdateMatch(match *models.Match) error {
	// Check if match exists
	existingMatch, err := s.repo.FindByID(match.ID)
	if err != nil {
		return errors.New("match not found")
	}

	// Prevent updating finished matches
	if existingMatch.Status == models.Finished {
		return errors.New("cannot update a finished match")
	}

	return s.repo.Update(match)
}

func (s *matchService) DeleteMatch(id uint) error {
	return s.repo.Delete(id)
}

func (s *matchService) ReportMatchResult(matchID uint, homeScore, awayScore int, goals []models.Goal) error {
	match, err := s.repo.FindByID(matchID)
	if err != nil {
		return errors.New("match not found")
	}

	// Update match scores and status
	homeScoreInt := homeScore
	awayScoreInt := awayScore
	match.HomeScore = &homeScoreInt
	match.AwayScore = &awayScoreInt
	match.Status = models.Finished

	// Use repository's transaction support
	return s.repo.WithTransaction(func(repo repositories.MatchRepository) error {
		// Update match
		if err := repo.Update(match); err != nil {
			return err
		}

		// Delete existing goals
		db := repo.GetDB()
		if err := db.Where("match_id = ?", matchID).Delete(&models.Goal{}).Error; err != nil {
			return err
		}

		// Add new goals
		for _, goal := range goals {
			goal.MatchID = matchID
			if err := db.Create(&goal).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

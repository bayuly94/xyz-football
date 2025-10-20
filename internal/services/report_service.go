package services

import (
	"sort"
	"time"

	"xyz-football/internal/models"
	"xyz-football/internal/repositories"
)

type ReportService interface {
	GetStandings() ([]TeamStanding, error)
	GetTopScorers(limit int) ([]PlayerGoals, error)
	GetMatchReport(matchID uint) (*MatchReport, error)
}

type reportService struct {
	repo     repositories.MatchRepository
	teamRepo repositories.TeamRepository
}

type TeamStanding struct {
	TeamID    uint   `json:"team_id"`
	TeamName  string `json:"team_name"`
	Played    int    `json:"played"`
	Won       int    `json:"won"`
	Drawn     int    `json:"drawn"`
	Lost      int    `json:"lost"`
	GoalsFor  int    `json:"goals_for"`
	GoalsAway int    `json:"goals_away"`
	Points    int    `json:"points"`
}

type PlayerGoals struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
	TeamName   string `json:"team_name"`
	Goals      int    `json:"goals"`
}

type MatchReport struct {
	MatchID    uint          `json:"match_id"`
	HomeTeam   string        `json:"home_team"`
	AwayTeam   string        `json:"away_team"`
	HomeScore  *int          `json:"home_score,omitempty"`
	AwayScore  *int          `json:"away_score,omitempty"`
	MatchTime  string        `json:"match_time"`
	Status     string        `json:"status"`
	Goals      []Goal        `json:"goals,omitempty"`
	TopScorers []PlayerGoals `json:"top_scorers,omitempty"`
}

type Goal struct {
	PlayerName string `json:"player_name"`
	Minute     int    `json:"minute"`
	IsOwnGoal  bool   `json:"is_own_goal"`
}

func NewReportService(matchRepo repositories.MatchRepository, teamRepo repositories.TeamRepository) ReportService {
	return &reportService{
		repo:     matchRepo,
		teamRepo: teamRepo,
	}
}

func (s *reportService) GetStandings() ([]TeamStanding, error) {
	teams, err := s.teamRepo.FindAll()
	if err != nil {
		return nil, err
	}

	standings := make(map[uint]*TeamStanding)

	// Initialize standings for all teams
	for _, team := range teams {
		standings[team.ID] = &TeamStanding{
			TeamID:    team.ID,
			TeamName:  team.Name,
			Played:    0,
			Won:       0,
			Drawn:     0,
			Lost:      0,
			GoalsFor:  0,
			GoalsAway: 0,
			Points:    0,
		}
	}

	// Process all finished matches
	matches, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if match.Status != models.Finished || match.HomeScore == nil || match.AwayScore == nil {
			continue
		}

		home := standings[match.HomeTeamID]
		away := standings[match.AwayTeamID]

		home.Played++
		away.Played++

		home.GoalsFor += *match.HomeScore
		home.GoalsAway += *match.AwayScore
		away.GoalsFor += *match.AwayScore
		away.GoalsAway += *match.HomeScore

		switch {
		case *match.HomeScore > *match.AwayScore:
			home.Won++
			away.Lost++
			home.Points += 3
		case *match.HomeScore < *match.AwayScore:
			away.Won++
			home.Lost++
			away.Points += 3
		default:
			home.Drawn++
			away.Drawn++
			home.Points++
			away.Points++
		}
	}

	// Convert map to slice and sort by points, then goal difference, then goals for
	result := make([]TeamStanding, 0, len(standings))
	for _, standing := range standings {
		result = append(result, *standing)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Points != result[j].Points {
			return result[i].Points > result[j].Points
		}
		iGD := result[i].GoalsFor - result[i].GoalsAway
		jGD := result[j].GoalsFor - result[j].GoalsAway
		if iGD != jGD {
			return iGD > jGD
		}
		return result[i].GoalsFor > result[j].GoalsFor
	})

	return result, nil
}

func (s *reportService) GetTopScorers(limit int) ([]PlayerGoals, error) {
	// count goals by player across all matches
	type row struct {
		PlayerID   uint
		Goals      int64
		PlayerName string
		TeamName   string
	}

	var rows []row
	err := s.repo.GetDB().Table("goals").
		Select(
			"goals.player_id as player_id,"+
			"COUNT(*) as goals,"+
			"players.name as player_name,"+
			"teams.name as team_name").
		Joins("JOIN players ON players.id = goals.player_id").
		Joins("JOIN teams ON teams.id = players.team_id").
		Group("goals.player_id, players.name, teams.name").
		Order("goals DESC, players.name ASC").
		Limit(limit).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	// Convert to PlayerGoals slice
	var result []PlayerGoals
	for _, row := range rows {
		result = append(result, PlayerGoals{
			PlayerID:   row.PlayerID,
			PlayerName: row.PlayerName,
			TeamName:   row.TeamName,
			Goals:      int(row.Goals),
		})
	}

	// If no results, return empty slice instead of nil
	if result == nil {
		result = []PlayerGoals{}
	}

	return result, nil
}

func (s *reportService) GetMatchReport(matchID uint) (*MatchReport, error) {
	match, err := s.repo.FindByID(matchID)
	if err != nil {
		return nil, err
	}

	report := &MatchReport{
		MatchID:   match.ID,
		HomeTeam:  match.HomeTeam.Name,
		AwayTeam:  match.AwayTeam.Name,
		HomeScore: match.HomeScore,
		AwayScore: match.AwayScore,
		MatchTime: match.MatchTime.Format(time.RFC3339),
		Status:    string(match.Status),
	}

	// Add goals to report
	for _, goal := range match.Goals {
		report.Goals = append(report.Goals, Goal{
			PlayerName: goal.Player.Name,
			Minute:     goal.Minute,
		})
	}

	return report, nil
}

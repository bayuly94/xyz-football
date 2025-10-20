package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ================== MODELS ==================

type Team struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" binding:"required"`
	LogoURL     string    `json:"logo_url"`
	FoundedYear int       `json:"founded_year"`
	StadiumAddr string    `json:"stadium_address"`
	City        string    `json:"city"`
	Players     []Player  `json:"players,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlayerPosition string

const (
	Striker    PlayerPosition = "penyerang"
	Midfield   PlayerPosition = "gelandang"
	Defender   PlayerPosition = "bertahan"
	Goalkeeper PlayerPosition = "penjaga gawang"
)

type Player struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	TeamID    uint           `json:"team_id" binding:"required"`
	Name      string         `json:"name" binding:"required"`
	HeightCM  float64        `json:"height_cm"`
	WeightKG  float64        `json:"weight_kg"`
	Position  PlayerPosition `json:"position" binding:"required,oneof=penyerang gelandang bertahan 'penjaga gawang'"`
	Number    int            `json:"number" binding:"required,min=1,max=99"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`

	Team Team `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Unique number per team enforced by composite unique index
func (Player) TableName() string { return "players" }

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

	HomeTeam Team `json:"-" gorm:"foreignKey:HomeTeamID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	AwayTeam Team `json:"-" gorm:"foreignKey:AwayTeamID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Goal struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MatchID   uint      `json:"match_id" binding:"required"`
	PlayerID  uint      `json:"player_id" binding:"required"`
	Minute    int       `json:"minute" binding:"required,min=0,max=130"`
	CreatedAt time.Time `json:"created_at"`

	Match  Match  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Player Player `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

// ================== DB INIT ==================

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("football.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// Migrations
	err = db.AutoMigrate(&Team{}, &Player{}, &Match{}, &Goal{})
	if err != nil {
		log.Fatal("migration error:", err)
	}

	// Composite unique index: team_id + number
	err = db.SetupJoinTable(&Match{}, "Goals", &Goal{})
	if err != nil {
		log.Println("join table setup:", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_players_team_number ON players(team_id, number);`).Error; err != nil {
		log.Fatal("failed to create unique index:", err)
	}
}

// ================== REQUEST DTOs ==================

type CreateTeamDTO struct {
	Name        string `json:"name" binding:"required"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	StadiumAddr string `json:"stadium_address"`
	City        string `json:"city"`
}

type CreatePlayerDTO struct {
	TeamID   uint           `json:"team_id" binding:"required"`
	Name     string         `json:"name" binding:"required"`
	HeightCM float64        `json:"height_cm"`
	WeightKG float64        `json:"weight_kg"`
	Position PlayerPosition `json:"position" binding:"required,oneof=penyerang gelandang bertahan 'penjaga gawang'"`
	Number   int            `json:"number" binding:"required,min=1,max=99"`
}

type ScheduleMatchDTO struct {
	MatchTime  time.Time `json:"match_time" binding:"required"` // RFC3339: "2025-10-16T15:00:00+07:00"
	HomeTeamID uint      `json:"home_team_id" binding:"required"`
	AwayTeamID uint      `json:"away_team_id" binding:"required,nefield=HomeTeamID"`
}

type ReportGoalDTO struct {
	PlayerID uint `json:"player_id" binding:"required"`
	Minute   int  `json:"minute" binding:"required,min=0,max=130"`
}

type ReportResultDTO struct {
	HomeScore int             `json:"home_score" binding:"required,min=0"`
	AwayScore int             `json:"away_score" binding:"required,min=0"`
	Goals     []ReportGoalDTO `json:"goals" binding:"dive"`
}

// ================== RESPONSE DTOs ==================

type MatchReportItem struct {
	MatchID                uint      `json:"match_id"`
	MatchTime              time.Time `json:"match_time"`
	HomeTeam               string    `json:"home_team"`
	AwayTeam               string    `json:"away_team"`
	FinalScore             string    `json:"final_score"`  // e.g. "2-1" (0-0 if belum ada skor)
	StatusAkhir            string    `json:"status_akhir"` // "Tim Home Menang" / "Tim Away Menang" / "Draw" / "Scheduled"
	TopScorerPlayer        *string   `json:"pencetak_gol_terbanyak,omitempty"`
	HomeWinsUntilThisMatch int       `json:"home_wins_until_this_match"`
	AwayWinsUntilThisMatch int       `json:"away_wins_until_this_match"`
}

// ================== HELPERS ==================

func statusAkhirText(m Match) string {
	if m.Status != Finished || m.HomeScore == nil || m.AwayScore == nil {
		return "Scheduled"
	}
	if *m.HomeScore > *m.AwayScore {
		return "Tim Home Menang"
	}
	if *m.HomeScore < *m.AwayScore {
		return "Tim Away Menang"
	}
	return "Draw"
}

func topScorerName(matchID uint) (*string, error) {
	// count goals by player for match
	type row struct {
		PlayerID uint
		Count    int64
		Name     string
	}
	var rows []row
	err := db.Table("goals").
		Select("player_id, COUNT(*) as count, players.name as name").
		Joins("JOIN players ON players.id = goals.player_id").
		Where("goals.match_id = ?", matchID).
		Group("player_id, players.name").
		Order("count DESC, players.name ASC").
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0].Name, nil
}

func cumulativeWinsUntil(match Match, teamID uint) (int, error) {
	var count int64
	// Semua match finished dengan match_time <= current
	err := db.Model(&Match{}).
		Where("status = ?", Finished).
		Where("match_time <= ?", match.MatchTime).
		Where(
			db.Where("home_team_id = ? AND home_score > away_score", teamID).
				Or("away_team_id = ? AND away_score > home_score", teamID),
		).
		Count(&count).Error
	return int(count), err
}

// ================== HANDLERS ==================

// ---- Teams ----
func listTeams(c *gin.Context) {
	var teams []Team
	if err := db.Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}
	c.JSON(http.StatusOK, teams)
}

func createTeam(c *gin.Context) {
	var dto CreateTeamDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := Team{
		Name:        dto.Name,
		LogoURL:     dto.LogoURL,
		FoundedYear: dto.FoundedYear,
		StadiumAddr: dto.StadiumAddr,
		City:        dto.City,
	}
	if err := db.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func getTeam(c *gin.Context) {
	var team Team
	if err := db.Preload("Players").First(&team, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	c.JSON(http.StatusOK, team)
}

func updateTeam(c *gin.Context) {
	var team Team
	if err := db.First(&team, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	var dto CreateTeamDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team.Name = dto.Name
	team.LogoURL = dto.LogoURL
	team.FoundedYear = dto.FoundedYear
	team.StadiumAddr = dto.StadiumAddr
	team.City = dto.City
	if err := db.Save(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team"})
		return
	}
	c.JSON(http.StatusOK, team)
}

func deleteTeam(c *gin.Context) {
	if err := db.Delete(&Team{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Players ----
func listPlayers(c *gin.Context) {
	var players []Player
	q := db.Preload("Team")
	if teamID := c.Query("team_id"); teamID != "" {
		q = q.Where("team_id = ?", teamID)
	}
	if err := q.Find(&players).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list players"})
		return
	}
	c.JSON(http.StatusOK, players)
}

func createPlayer(c *gin.Context) {
	var dto CreatePlayerDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Pastikan tim ada
	var t Team
	if err := db.First(&t, dto.TeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team not found"})
		return
	}
	p := Player{
		TeamID:   dto.TeamID,
		Name:     dto.Name,
		HeightCM: dto.HeightCM,
		WeightKG: dto.WeightKG,
		Position: dto.Position,
		Number:   dto.Number,
	}
	if err := db.Create(&p).Error; err != nil {
		// kemungkinan unique violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nomor punggung sudah dipakai di tim ini"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "gagal membuat pemain (cek nomor unik/validasi)"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func getPlayer(c *gin.Context) {
	var p Player
	if err := db.First(&p, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func updatePlayer(c *gin.Context) {
	var p Player
	if err := db.First(&p, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	var dto CreatePlayerDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Pastikan tim baru ada (kalau dipindah)
	var t Team
	if err := db.First(&t, dto.TeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "team not found"})
		return
	}
	p.TeamID = dto.TeamID
	p.Name = dto.Name
	p.HeightCM = dto.HeightCM
	p.WeightKG = dto.WeightKG
	p.Position = dto.Position
	p.Number = dto.Number

	if err := db.Save(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nomor punggung sudah dipakai di tim ini"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "gagal mengubah pemain (cek nomor unik/validasi)"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func deletePlayer(c *gin.Context) {
	if err := db.Delete(&Player{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete player"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Matches (Jadwal) ----
func listMatches(c *gin.Context) {
	var matches []Match
	if err := db.Preload("Goals").
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("match_time ASC").
		Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list matches"})
		return
	}
	c.JSON(http.StatusOK, matches)
}

func scheduleMatch(c *gin.Context) {
	var dto ScheduleMatchDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate teams exist
	var ht, at Team
	if err := db.First(&ht, dto.HomeTeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "home team not found"})
		return
	}
	if err := db.First(&at, dto.AwayTeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "away team not found"})
		return
	}
	m := Match{
		MatchTime:  dto.MatchTime,
		HomeTeamID: dto.HomeTeamID,
		AwayTeamID: dto.AwayTeamID,
		Status:     Scheduled,
	}
	if err := db.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule match"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func getMatch(c *gin.Context) {
	var m Match
	if err := db.Preload("Goals").
		Preload("HomeTeam").
		Preload("AwayTeam").
		First(&m, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func updateMatch(c *gin.Context) {
	var m Match
	if err := db.First(&m, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	var dto ScheduleMatchDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate teams
	var ht, at Team
	if err := db.First(&ht, dto.HomeTeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "home team not found"})
		return
	}
	if err := db.First(&at, dto.AwayTeamID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "away team not found"})
		return
	}
	// Only allow update while scheduled
	if m.Status == Finished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update finished match"})
		return
	}
	m.MatchTime = dto.MatchTime
	m.HomeTeamID = dto.HomeTeamID
	m.AwayTeamID = dto.AwayTeamID
	if err := db.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func deleteMatch(c *gin.Context) {
	if err := db.Delete(&Match{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete match"})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Report Result ----
func reportResult(c *gin.Context) {
	var m Match
	if err := db.Preload("Goals").First(&m, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	if m.Status == Finished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "match already finished"})
		return
	}

	var dto ReportResultDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate scorers belong to either team in the match
	if len(dto.Goals) > 0 {
		var playerIDs []uint
		for _, g := range dto.Goals {
			playerIDs = append(playerIDs, g.PlayerID)
		}
		var players []Player
		if err := db.Where("id IN ?", playerIDs).Find(&players).Error; err != nil || len(players) != len(playerIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more players not found"})
			return
		}
		for _, p := range players {
			if p.TeamID != m.HomeTeamID && p.TeamID != m.AwayTeamID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "goal scorer must belong to home or away team"})
				return
			}
		}
	}

	// Transaction: upsert scores + replace goals
	err := db.Transaction(func(tx *gorm.DB) error {
		// replace goals for this match
		if err := tx.Where("match_id = ?", m.ID).Delete(&Goal{}).Error; err != nil {
			return err
		}
		if len(dto.Goals) > 0 {
			var goals []Goal
			for _, g := range dto.Goals {
				goals = append(goals, Goal{
					MatchID:  m.ID,
					PlayerID: g.PlayerID,
					Minute:   g.Minute,
				})
			}
			if err := tx.Create(&goals).Error; err != nil {
				return err
			}
		}
		// update score + status
		m.HomeScore = &dto.HomeScore
		m.AwayScore = &dto.AwayScore
		m.Status = Finished
		if err := tx.Save(&m).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to report result"})
		return
	}

	// return updated match
	if err := db.Preload("Goals").First(&m, m.ID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "result saved"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// ---- Reports ----
func matchReportItem(m Match) (MatchReportItem, error) {
	// preload team names
	var home, away Team
	if err := db.First(&home, m.HomeTeamID).Error; err != nil {
		return MatchReportItem{}, err
	}
	if err := db.First(&away, m.AwayTeamID).Error; err != nil {
		return MatchReportItem{}, err
	}

	final := "0-0"
	if m.HomeScore != nil && m.AwayScore != nil {
		final = formatScore(*m.HomeScore, *m.AwayScore)
	}

	top, err := topScorerName(m.ID)
	if err != nil {
		return MatchReportItem{}, err
	}

	homeWins, err := cumulativeWinsUntil(m, m.HomeTeamID)
	if err != nil {
		return MatchReportItem{}, err
	}
	awayWins, err := cumulativeWinsUntil(m, m.AwayTeamID)
	if err != nil {
		return MatchReportItem{}, err
	}

	return MatchReportItem{
		MatchID:                m.ID,
		MatchTime:              m.MatchTime,
		HomeTeam:               home.Name,
		AwayTeam:               away.Name,
		FinalScore:             final,
		StatusAkhir:            statusAkhirText(m),
		TopScorerPlayer:        top,
		HomeWinsUntilThisMatch: homeWins,
		AwayWinsUntilThisMatch: awayWins,
	}, nil
}

func formatScore(h, a int) string {
	return fmtInt(h) + "-" + fmtInt(a)
}

func fmtInt(n int) string {
	return strconvItoa(n)
}

// Lightweight strconv to avoid extra import noise in the big snippet
func strconvItoa(i int) string {
	// standard library strconv would be simpler; keeping minimal deps in this snippet
	if i == 0 {
		return "0"
	}
	sign := ""
	if i < 0 {
		sign = "-"
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return sign + string(b[pos:])
}

func listReports(c *gin.Context) {
	var matches []Match
	if err := db.Preload(clause.Associations).
		Order("match_time ASC").
		Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query matches"})
		return
	}
	resp := make([]MatchReportItem, 0, len(matches))
	for _, m := range matches {
		item, err := matchReportItem(m)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build report"})
			return
		}
		resp = append(resp, item)
	}
	c.JSON(http.StatusOK, resp)
}

func getReportByMatch(c *gin.Context) {
	var m Match
	if err := db.First(&m, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}
	item, err := matchReportItem(m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build report"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// ================== ROUTER & MAIN ==================

func main() {
	initDB()

	r := gin.Default()

	api := r.Group("/api")
	{
		// Teams
		api.GET("/teams", listTeams)
		api.POST("/teams", createTeam)
		api.GET("/teams/:id", getTeam)
		api.PUT("/teams/:id", updateTeam)
		api.DELETE("/teams/:id", deleteTeam)

		// Players
		api.GET("/players", listPlayers) // ?team_id=1
		api.POST("/players", createPlayer)
		api.GET("/players/:id", getPlayer)
		api.PUT("/players/:id", updatePlayer)
		api.DELETE("/players/:id", deletePlayer)

		// Matches (jadwal)
		api.GET("/matches", listMatches)
		api.POST("/matches", scheduleMatch)
		api.GET("/matches/:id", getMatch)
		api.PUT("/matches/:id", updateMatch)
		api.DELETE("/matches/:id", deleteMatch)

		// Report match result
		api.POST("/matches/:id/report", reportResult)

		// Reports
		api.GET("/reports/matches", listReports)
		api.GET("/reports/matches/:id", getReportByMatch)
	}

	log.Println("Server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

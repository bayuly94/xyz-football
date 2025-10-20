package handlers

import (
	"net/http"
	"strconv"
	"time"

	"xyz-football/internal/models"
	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
)

type MatchHandler struct {
	service services.MatchService
}

func NewMatchHandler(service services.MatchService) *MatchHandler {
	return &MatchHandler{service: service}
}

type CreateMatchRequest struct {
	MatchTime  time.Time `json:"match_time" binding:"required"`
	HomeTeamID uint      `json:"home_team_id" binding:"required"`
	AwayTeamID uint      `json:"away_team_id" binding:"required,nefield=HomeTeamID"`
}

type ReportGoalRequest struct {
	PlayerID uint `json:"player_id" binding:"required"`
	Minute   int  `json:"minute" binding:"required,min=1,max=130"`
}

type ReportResultRequest struct {
	HomeScore int                 `json:"home_score"`
	AwayScore int                 `json:"away_score"`
	Goals     []ReportGoalRequest `json:"goals"`
}

func (h *MatchHandler) Create(c *gin.Context) {
	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	match := &models.Match{
		MatchTime:  req.MatchTime,
		HomeTeamID: req.HomeTeamID,
		AwayTeamID: req.AwayTeamID,
		Status:     models.Scheduled,
	}

	if err := h.service.CreateMatch(match); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Match created successfully",
		"data":    match,
	})
}

func (h *MatchHandler) List(c *gin.Context) {
	// Check for date range query params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate != "" && endDate != "" {
		start, err1 := time.Parse(time.RFC3339, startDate)
		end, err2 := time.Parse(time.RFC3339, endDate)

		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use RFC3339"})
			return
		}

		matches, err := h.service.GetMatchesByDateRange(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": matches,
		})
		return
	}

	// If no date range, get all matches
	matches, err := h.service.GetAllMatches()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": matches,
	})
}

func (h *MatchHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	match, err := h.service.GetMatchByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": match,
	})
}

func (h *MatchHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	var req CreateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	match := &models.Match{
		ID:         uint(id),
		MatchTime:  req.MatchTime,
		HomeTeamID: req.HomeTeamID,
		AwayTeamID: req.AwayTeamID,
	}

	if err := h.service.UpdateMatch(match); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Match updated successfully",
		"data":    match,
	})
}

func (h *MatchHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	if err := h.service.DeleteMatch(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete match"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MatchHandler) ReportResult(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	var req ReportResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert request goals to models.Goal
	goals := make([]models.Goal, len(req.Goals))
	for i, g := range req.Goals {
		goals[i] = models.Goal{
			PlayerID: g.PlayerID,
			Minute:   g.Minute,
		}
	}

	if err := h.service.ReportMatchResult(uint(matchID), req.HomeScore, req.AwayScore, goals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match result reported successfully"})
}

func (h *MatchHandler) GetByTeam(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("teamId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	matches, err := h.service.GetMatchesByTeam(uint(teamID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": matches,
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"xyz-football/internal/models"
	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	service services.TeamService
}

func NewTeamHandler(service services.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	StadiumAddr string `json:"stadium_address"`
	City        string `json:"city"`
}

func (h *TeamHandler) Create(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := &models.Team{
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		FoundedYear: req.FoundedYear,
		StadiumAddr: req.StadiumAddr,
		City:        req.City,
	}

	if err := h.service.CreateTeam(team); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"data":    team,
	})
}

func (h *TeamHandler) List(c *gin.Context) {
	teams, err := h.service.GetAllTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": teams,
	})
}

func (h *TeamHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	team, err := h.service.GetTeamByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": team,
	})
}

func (h *TeamHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := &models.Team{
		ID:          uint(id),
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		FoundedYear: req.FoundedYear,
		StadiumAddr: req.StadiumAddr,
		City:        req.City,
	}

	// check if exist
	if _, err := h.service.GetTeamByID(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if err := h.service.UpdateTeam(team); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team updated successfully",
		"data":    team,
	})
}

func (h *TeamHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	// check if exist
	if _, err := h.service.GetTeamByID(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if err := h.service.DeleteTeam(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Team deleted successfully",
	})
}

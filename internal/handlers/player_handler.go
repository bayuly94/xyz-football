package handlers

import (
	"net/http"
	"strconv"

	"xyz-football/internal/models"
	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	service services.PlayerService
}

func NewPlayerHandler(service services.PlayerService) *PlayerHandler {
	return &PlayerHandler{service: service}
}

type CreatePlayerRequest struct {
	TeamID   uint                  `json:"team_id" binding:"required"`
	Name     string                `json:"name" binding:"required"`
	HeightCM float64               `json:"height_cm"`
	WeightKG float64               `json:"weight_kg"`
	Position models.PlayerPosition `json:"position" binding:"required,oneof=striker midfielder defender goalkeeper"`
	Number   int                   `json:"number" binding:"required,min=1,max=99"`
}

func (h *PlayerHandler) Create(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := &models.Player{
		TeamID:   req.TeamID,
		Name:     req.Name,
		HeightCM: req.HeightCM,
		WeightKG: req.WeightKG,
		Position: req.Position,
		Number:   req.Number,
	}

	if err := h.service.CreatePlayer(player); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Player created successfully",
		"data":    player,
	})
}

func (h *PlayerHandler) List(c *gin.Context) {
	players, err := h.service.GetAllPlayers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": players,
	})
}

func (h *PlayerHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}

	player, err := h.service.GetPlayerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": player,
	})
}

func (h *PlayerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}

	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := &models.Player{
		ID:       uint(id),
		TeamID:   req.TeamID,
		Name:     req.Name,
		HeightCM: req.HeightCM,
		WeightKG: req.WeightKG,
		Position: req.Position,
		Number:   req.Number,
	}

	if err := h.service.UpdatePlayer(player); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player updated successfully",
		"data":    player,
	})
}

func (h *PlayerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}

	if err := h.service.DeletePlayer(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete player"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Player deleted successfully",
	})
}

func (h *PlayerHandler) ListByTeam(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("teamId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	players, err := h.service.GetPlayersByTeam(uint(teamID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": players,
	})
}

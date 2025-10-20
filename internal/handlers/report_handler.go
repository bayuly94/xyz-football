package handlers

import (
	"net/http"
	"strconv"

	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service services.ReportService
}

func NewReportHandler(service services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) GetStandings(c *gin.Context) {
	standings, err := h.service.GetStandings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch standings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": standings,
	})
}

func (h *ReportHandler) GetTopScorers(c *gin.Context) {
	limit := 10 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	scorers, err := h.service.GetTopScorers(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch top scorers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scorers,
	})
}

func (h *ReportHandler) GetMatchReport(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	report, err := h.service.GetMatchReport(uint(matchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": report,
	})
}

// Response helper functions
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

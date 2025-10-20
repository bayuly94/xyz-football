package handlers

import (
	"net/http"

	"xyz-football/internal/models"
	"xyz-football/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service services.AdminService
}

func NewAdminHandler(service services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Login handles admin login
// @Summary Admin login
// @Description Login with email and password
// @Tags admin
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, admin, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
		"data":    admin,
	})
}

// Register handles admin registration
// @Summary Register new admin
// @Description Register a new admin user
// @Tags admin
// @Accept json
// @Produce json
// @Param input body RegisterRequest true "Admin details"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/register [post]
func (h *AdminHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin := &models.Admin{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// check if admin exist not allowed
	if _, err := h.service.FindByEmail(req.Email); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin already exists. only 1 admin is allowed to register"})
		return
	}

	if err := h.service.Register(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "admin registered successfully",
	})
}

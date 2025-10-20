package services

import (
	"errors"
	"xyz-football/internal/models"
	"xyz-football/internal/repositories"
	"xyz-football/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
	Login(email, password string) (string, *models.Admin, error)
	Register(admin *models.Admin) error
	FindByEmail(email string) (*models.Admin, error)
}

type adminService struct {
	repo repositories.AdminRepository
}

func NewAdminService(repo repositories.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) Login(email, password string) (string, *models.Admin, error) {
	admin, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := generateJWT(admin.ID)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	admin.Password = "" // remove password from response

	return token, admin, nil
}

func (s *adminService) Register(admin *models.Admin) error {
	// Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	admin.Password = string(hashedPassword)
	return s.repo.Create(admin)
}

func (s *adminService) FindByEmail(email string) (*models.Admin, error) {
	return s.repo.FindByEmail(email)
}

// Helper function to generate JWT token
func generateJWT(userID uint) (string, error) {
	return utils.GenerateToken(userID)
}

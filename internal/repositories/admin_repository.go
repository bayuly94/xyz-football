package repositories

import (
	"xyz-football/internal/models"

	"gorm.io/gorm"
)

type AdminRepository interface {
	FindByEmail(email string) (*models.Admin, error)
	Create(admin *models.Admin) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) FindByEmail(email string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Create(admin *models.Admin) error {
	return r.db.Create(admin).Error
}

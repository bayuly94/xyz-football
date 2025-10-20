package database

import (
	"log"

	"xyz-football/internal/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Admin{},
		&models.Team{},
		&models.Player{},
		&models.Match{},
		&models.Goal{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("✅ Database migration complete")
}

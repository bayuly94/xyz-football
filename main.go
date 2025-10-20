package main

import (
	"log"

	"xyz-football/config"
	"xyz-football/internal/database"
	"xyz-football/internal/routers"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	database.Migrate(db)

	r := routers.Setup(db)

	log.Printf("Server running on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}

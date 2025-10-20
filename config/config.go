package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
	DBPath   string // for SQLite
	Port     string
}

func Load() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		DBDriver: getEnv("DB_DRIVER", "sqlite"),
		DBHost:   getEnv("DB_HOST", "localhost"),
		DBPort:   getEnv("DB_PORT", "3306"),
		DBUser:   getEnv("DB_USER", ""),
		DBPass:   getEnv("DB_PASS", ""),
		DBName:   getEnv("DB_NAME", "xyz_football"),
		DBPath:   getEnv("DB_PATH", "storage/xyz_football.db"),
		Port:     getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

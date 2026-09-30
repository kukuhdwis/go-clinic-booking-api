package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv         string
	AppPort        string
	AppName        string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	JWTSecret      string
	JWTExpireHours int
}

func LoadConfig() *Config {
	// Try loading .env if available, ignore error in production environments
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}

	expireHours, err := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	if err != nil {
		expireHours = 24
	}

	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		AppPort:        getEnv("APP_PORT", "8080"),
		AppName:        getEnv("APP_NAME", "ClinicEase"),
		DBHost:         getEnv("DB_HOST", "127.0.0.1"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "clinicease_db"),
		JWTSecret:      getEnv("JWT_SECRET", "default-clinic-secret-change-in-production"),
		JWTExpireHours: expireHours,
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

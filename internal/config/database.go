package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDatabase(cfg *Config) (*gorm.DB, error) {
	// Step 1: Ensure database exists
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)

	rawDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL server: %w", err)
	}
	defer rawDB.Close()

	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.DBName)
	if _, err := rawDB.Exec(createDBSQL); err != nil {
		return nil, fmt.Errorf("failed to create database %s: %w", cfg.DBName, err)
	}

	// Step 2: Connect to specific database with GORM
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Step 3: Connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve generic database handle: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("Successfully connected to MySQL database: %s\n", cfg.DBName)
	return db, nil
}

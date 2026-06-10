package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *Config) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=Asia/Jakarta",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
}

func AutoMigrate(models ...interface{}) {
	if err := DB.AutoMigrate(models...); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Auto-migration completed")
}

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

func EnsureTestDB() {
	cfg := Load()
	// Use GORM to check and create
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Printf("[testdb] cannot connect to postgres: %v", err)
		return
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var exists bool
	db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", "spendwise_test_db").Scan(&exists)
	if !exists {
		db.Exec("CREATE DATABASE spendwise_test_db")
		log.Println("[testdb] created spendwise_test_db")
	}
}

func AutoMigrate(models ...interface{}) {
	if err := DB.AutoMigrate(models...); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}
	log.Println("Auto-migration completed")
}

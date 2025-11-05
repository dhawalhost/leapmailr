package database

import (
	"fmt"
	"log"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() error {
	conf := config.GetConfig()

	// Build DSN from config
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		conf.DBHost,
		conf.DBUser,
		conf.DBPassword,
		conf.DBName,
		conf.DBPort,
		conf.DBSSLMode,
		// conf.DBTimezone,
	)

	// Set log level based on environment
	logLevel := logger.Silent
	// if conf.EnvMode != "release" {
	// 	logLevel = logger.Info
	// }

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
	// 	Logger: logger.Default.LogMode(logLevel),
	// })
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db

	// Run auto migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// runMigrations runs all database migrations
func runMigrations() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.UserOrganization{},
		&models.AuthToken{},
		&models.APIKey{},
		&models.APIKeyPair{},
		&models.APIKeyUsageLog{},
		&models.UserSession{},
		&models.AuditLog{},
		&models.PasswordHistory{},
		&models.Project{},
		&models.EmailService{},
		&models.Template{},
		&models.EmailLog{},
		&models.WebhookEvent{},
		&models.CaptchaConfig{},
		&models.Suppression{},
		&models.AutoReplyConfig{},
		&models.AutoReplyLog{},
		&models.Contact{},
		&models.ContactList{},
		// Email tracking models
		&models.EmailTracking{},
		&models.EmailOpenEvent{},
		&models.EmailClickEvent{},
	)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// CloseDB closes the database connection
func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

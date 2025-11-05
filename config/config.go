package config

import (
	"github.com/spf13/viper"
)

var (
	appConfig *AppConfig
)

type AppConfig struct {
	Port              string `mapstructure:"PORT"`
	DefaultSenderMail string `mapstructure:"DEFAULT_SENDER_MAIL"`
	CompanyName       string `mapstructure:"COMPANY_NAME"`
	EnvMode           string `mapstructure:"ENV_MODE"`
	ContactMail       string `mapstructure:"CONTACT_MAIL"`
	LogoURL           string `mapstructure:"LOGO_URL"`

	// Database Configuration
	DBHost     string `mapstructure:"DB_HOST"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPort     int    `mapstructure:"DB_PORT"`
	DBSSLMode  string `mapstructure:"DB_SSLMODE"`
	DBTimezone string `mapstructure:"DB_TIMEZONE"`

	// JWT Configuration
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTExpirationHours int    `mapstructure:"JWT_EXPIRATION_HOURS"`
	JWTRefreshDays     int    `mapstructure:"JWT_REFRESH_DAYS"`

	// Encryption
	EncryptionKey string `mapstructure:"ENCRYPTION_KEY"`

	// Rate Limiting
	RateLimit int `mapstructure:"RATE_LIMIT"`

	// External Services
	NRLicense string `mapstructure:"NR_LICENSE_KEY"`
}

// Load configuration from environment file using Viper
func LoadConfig() *AppConfig {
	viper.SetConfigFile(".env")
	// Try to read .env file, but don't panic if it doesn't exist
	// Environment variables will be used instead (e.g., in Docker)
	_ = viper.ReadInConfig() // Ignore error if .env doesn't exist
	viper.AutomaticEnv()
	appConfig = &AppConfig{
		Port:              viper.GetString("PORT"),
		DefaultSenderMail: viper.GetString("DEFAULT_SENDER_MAIL"),
		CompanyName:       viper.GetString("COMPANY_NAME"),
		EnvMode:           viper.GetString("ENV_MODE"),
		ContactMail:       viper.GetString("CONTACT_MAIL"),
		LogoURL:           viper.GetString("LOGO_URL"),

		// Database
		DBHost:     viper.GetString("DB_HOST"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBPort:     viper.GetInt("DB_PORT"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),
		DBTimezone: viper.GetString("DB_TIMEZONE"),

		// JWT
		JWTSecret:          viper.GetString("JWT_SECRET"),
		JWTExpirationHours: viper.GetInt("JWT_EXPIRATION_HOURS"),
		JWTRefreshDays:     viper.GetInt("JWT_REFRESH_DAYS"),

		// Encryption
		EncryptionKey: viper.GetString("ENCRYPTION_KEY"),

		// Rate Limiting & Services
		RateLimit: viper.GetInt("RATE_LIMIT"),
		NRLicense: viper.GetString("NR_LICENSE_KEY"),
	}
	return appConfig
}

func GetConfig() *AppConfig {
	return appConfig
}

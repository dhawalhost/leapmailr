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

	SMTPServer string `mapstructure:"SMTP_SERVER_URL"`
	SMTPMail   string `mapstructure:"SMTP_USER"`
	SMTPSecret string `mapstructure:"SMTP_SECRET"`
	SMTPPort   int    `mapstructure:"SMTP_PORT"`
	RateLimit  int    `mapstructure:"RATE_LIMIT"`
	NRLicense  string `mapstructure:"NR_LICENSE_KEY"`
}

// Load configuration from environment file using Viper
func LoadConfig() *AppConfig {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	viper.AutomaticEnv()
	return appConfig
}

func GetConfig() *AppConfig {
	return appConfig
}

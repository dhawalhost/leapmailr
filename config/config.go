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
	appConfig = &AppConfig{
		Port:              viper.GetString("PORT"),
		DefaultSenderMail: viper.GetString("DEFAULT_SENDER_MAIL"),
		CompanyName:       viper.GetString("COMPANY_NAME"),
		EnvMode:           viper.GetString("ENV_MODE"),
		ContactMail:       viper.GetString("CONTACT_MAIL"),
		LogoURL:           viper.GetString("LOGO_URL"),
		SMTPServer:        viper.GetString("SMTP_SERVER_URL"),
		SMTPMail:          viper.GetString("SMTP_USER"),
		SMTPSecret:        viper.GetString("SMTP_SECRET"),
		SMTPPort:          viper.GetInt("SMTP_PORT"),
		RateLimit:         viper.GetInt("RATE_LIMIT"),
		NRLicense:         viper.GetString("NR_LICENSE_KEY"),
	}
	return appConfig
}

func GetConfig() *AppConfig {
	return appConfig
}

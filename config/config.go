package config

import (
	"github.com/spf13/viper"
)

var (
	appConfig AppConfig
)

type AppConfig struct {
	Port              string
	DefaultSenderMail string
	CompanyName       string
	EnvMode           string
	ContactMail       string
	LogoURL           string

	SMTPServer string
	SMTPMail   string
	SMTPSecret string
	SMTPPort   int
	RateLimit  int
}

var (
	LogoURL = "https://dhawalhost.com/img/general/dhlogov.png"
	MailTo  = "dhawalhost@gmail.com"
)

// Load configuration from environment file using Viper
func LoadConfig() AppConfig {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	appConfig.Port = viper.GetString("PORT")
	appConfig.DefaultSenderMail = viper.GetString("DEFAULT_SENDER_MAIL")
	appConfig.CompanyName = viper.GetString("COMPANY_NAME")
	appConfig.EnvMode = viper.GetString("ENV_MODE")
	appConfig.ContactMail = viper.GetString("CONTACT_MAIL")
	appConfig.LogoURL = viper.GetString("LOGO_URL")
	appConfig.SMTPServer = viper.GetString("SMTP_SERVER_URL")
	appConfig.SMTPMail = viper.GetString("SMTP_USER")
	appConfig.SMTPSecret = viper.GetString("SMTP_SECRET")
	appConfig.SMTPPort = viper.GetInt("SMTP_PORT")
	appConfig.RateLimit = viper.GetInt("RATE_LIMIT")
	return appConfig
}

func GetConfig() AppConfig {
	return appConfig
}

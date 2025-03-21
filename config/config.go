package config

import (
	"github.com/spf13/viper"
)

var (
	appConfig *AppConfig
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
	NRLicense  string
}

var (
	LogoURL = "https://dhawalhost.com/img/general/dhlogov.png"
	MailTo  = "dhawalhost@gmail.com"
)

// Load configuration from environment file using Viper
func LoadConfig() *AppConfig {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
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

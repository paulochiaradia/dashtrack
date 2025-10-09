package config

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	// Database
	DBSource string `mapstructure:"DB_SOURCE"`

	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`
	ServerEnv  string `mapstructure:"SERVER_ENV"`

	// JWT
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	JWTAccessExpireMinutes int    `mapstructure:"JWT_ACCESS_EXPIRE_MINUTES"`
	JWTRefreshExpireHours  int    `mapstructure:"JWT_REFRESH_EXPIRE_HOURS"`

	// Email
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	FromEmail    string `mapstructure:"FROM_EMAIL"`

	// Application
	AppName    string `mapstructure:"APP_NAME"`
	AppVersion string `mapstructure:"APP_VERSION"`
	AppURL     string `mapstructure:"APP_URL"`

	// Security
	BcryptCost               int `mapstructure:"BCRYPT_COST"`
	PasswordResetExpireHours int `mapstructure:"PASSWORD_RESET_EXPIRE_HOURS"`
}

var (
	once   sync.Once
	config *Config
)

func LoadConfig() *Config {
	once.Do(func() {
		// Load .env file
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}

		viper.AutomaticEnv()
		// Set defaults
		viper.SetDefault("SERVER_PORT", "8080")
		viper.SetDefault("SERVER_ENV", "development")
		viper.SetDefault("JWT_ACCESS_EXPIRE_MINUTES", 60) // Aumentado para 60 minutos durante testes
		viper.SetDefault("JWT_REFRESH_EXPIRE_HOURS", 24)
		viper.SetDefault("SMTP_PORT", 587)
		viper.SetDefault("BCRYPT_COST", 12)
		viper.SetDefault("PASSWORD_RESET_EXPIRE_HOURS", 1)
		viper.SetDefault("APP_NAME", "Dashtrack API")
		viper.SetDefault("APP_VERSION", "1.0.0")

		config = &Config{
			DBSource:                 viper.GetString("DB_SOURCE"),
			ServerPort:               viper.GetString("SERVER_PORT"),
			ServerEnv:                viper.GetString("SERVER_ENV"),
			JWTSecret:                viper.GetString("JWT_SECRET"),
			JWTAccessExpireMinutes:   viper.GetInt("JWT_ACCESS_EXPIRE_MINUTES"),
			JWTRefreshExpireHours:    viper.GetInt("JWT_REFRESH_EXPIRE_HOURS"),
			SMTPHost:                 viper.GetString("SMTP_HOST"),
			SMTPPort:                 viper.GetInt("SMTP_PORT"),
			SMTPUsername:             viper.GetString("SMTP_USERNAME"),
			SMTPPassword:             viper.GetString("SMTP_PASSWORD"),
			FromEmail:                viper.GetString("FROM_EMAIL"),
			AppName:                  viper.GetString("APP_NAME"),
			AppVersion:               viper.GetString("APP_VERSION"),
			AppURL:                   viper.GetString("APP_URL"),
			BcryptCost:               viper.GetInt("BCRYPT_COST"),
			PasswordResetExpireHours: viper.GetInt("PASSWORD_RESET_EXPIRE_HOURS"),
		}

		// Validate required fields
		if config.DBSource == "" {
			log.Fatal("DB_SOURCE is required")
		}
		if config.JWTSecret == "" {
			log.Fatal("JWT_SECRET is required")
		}
	})
	return config
}

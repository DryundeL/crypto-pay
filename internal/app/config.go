package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	App      AppConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type AppConfig struct {
	Env          string
	Port         string
	URL          string
	Version      string
	JWTSecret    string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	HelpMail     string
}

func LoadConfig() (*Config, error) {
	envPath := filepath.Join(".", ".env")
	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load()
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "crypto_pay"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		App: AppConfig{
			Env:          getEnv("APP_ENV", "development"),
			Port:         getEnv("APP_PORT", "8080"),
			URL:          getEnv("APP_URL", "http://localhost:8080"),
			Version:      getEnv("APP_VERSION", getBuildVersion()),
			JWTSecret:    getEnv("JWT_SECRET", ""),
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			SMTPFrom:     getEnv("SMTP_FROM", ""),
			HelpMail:     getEnv("HELP_MAIL", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if c.App.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if len(c.App.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	if c.App.Env != "development" && c.App.Env != "staging" && c.App.Env != "production" {
		return fmt.Errorf("APP_ENV must be one of: development, staging, production")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBuildVersion() string {
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}
	return fmt.Sprintf("%d", time.Now().Unix())
}

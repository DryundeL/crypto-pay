package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database      DatabaseConfig
	App           AppConfig
	Confirmations map[string]int
	EVM           EVMConfig
	Scanner       ScannerConfig
}

// EVMConfig holds deposit-address derivation settings for EVM networks.
type EVMConfig struct {
	// SepoliaXPub is an account-level extended public key at m/44'/60'/0'.
	// Child path used for deposits: m/44'/60'/0'/0/{index}.
	SepoliaXPub string
}

// ScannerConfig holds cmd/scanner poll settings.
type ScannerConfig struct {
	SepoliaRPCURL     string
	SepoliaStartBlock uint64
	PollInterval      time.Duration
	BlockBatch        int
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

	appCfg := AppConfig{
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
	}

	overrides, err := ParseConfirmationThresholds(getEnv("CONFIRMATION_THRESHOLDS", ""))
	if err != nil {
		return nil, fmt.Errorf("CONFIRMATION_THRESHOLDS: %w", err)
	}

	pollInterval, err := parseDurationEnv("SCANNER_POLL_INTERVAL", 3*time.Second)
	if err != nil {
		return nil, err
	}
	blockBatch, err := parsePositiveIntEnv("SCANNER_BLOCK_BATCH", 20)
	if err != nil {
		return nil, err
	}
	startBlock, err := parseUint64Env("EVM_SEPOLIA_START_BLOCK", 0)
	if err != nil {
		return nil, err
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
		App:           appCfg,
		Confirmations: mergeConfirmationThresholds(DefaultConfirmationThresholds(appCfg.Env), overrides),
		EVM: EVMConfig{
			SepoliaXPub: getEnv("EVM_SEPOLIA_XPUB", ""),
		},
		Scanner: ScannerConfig{
			SepoliaRPCURL:     getEnv("EVM_SEPOLIA_RPC_URL", ""),
			SepoliaStartBlock: startBlock,
			PollInterval:      pollInterval,
			BlockBatch:        blockBatch,
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

func parseDurationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return d, nil
}

func parsePositiveIntEnv(key string, defaultValue int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be an integer >= 1", key)
	}
	return n, nil
}

func parseUint64Env(key string, defaultValue uint64) (uint64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue, nil
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an unsigned integer", key)
	}
	return n, nil
}

func getBuildVersion() string {
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}
	return fmt.Sprintf("%d", time.Now().Unix())
}

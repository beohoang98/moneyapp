package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port            string
	DBPath          string
	JWTSecret       string
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOBucket     string
	MinIOUseSSL     bool
	BcryptCost      int
	TokenExpiryHours int
	Currency        string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DBPath:          getEnv("DB_PATH", "./moneyapp.db"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:     getEnv("MINIO_BUCKET", "moneyapp"),
		MinIOUseSSL:     getEnvBool("MINIO_USE_SSL", false),
		BcryptCost:      getEnvInt("BCRYPT_COST", 12),
		TokenExpiryHours: getEnvInt("TOKEN_EXPIRY_HOURS", 24),
		Currency:        getEnv("CURRENCY", "VND"),
	}

	if cfg.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET is not set. Using insecure default for development only.")
		cfg.JWTSecret = "dev-insecure-secret"
	} else if cfg.JWTSecret == "change-me" {
		log.Println("WARNING: JWT_SECRET is set to the example value. Change it in production.")
	}

	if cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH must not be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf("WARNING: invalid bool for %s=%q, using default %v", key, v, fallback)
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("WARNING: invalid int for %s=%q, using default %d", key, v, fallback)
		return fallback
	}
	return n
}

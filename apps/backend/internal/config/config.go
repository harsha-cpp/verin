package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	AppHost           string
	AppPort           string
	WebOrigin         string
	DatabaseURL       string
	RedisURL          string
	SessionSecret     string
	MFAEncryptionKey  string
	S3Bucket          string
	S3Region          string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3UsePathStyle    bool
	SignedURLTTL      time.Duration
	TikaEndpoint      string
	DefaultOrgID      string
	SessionCookieName string
	CSRFCookieName    string
	SessionTTL        time.Duration
	AuthPendingTTL    time.Duration
	LockoutThreshold  int32
	LockoutDuration   time.Duration
	SMTPHost          string
	SMTPPort          string
	SMTPFromAddress   string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppHost:           getEnv("APP_HOST", "0.0.0.0"),
		AppPort:           getEnv("APP_PORT", "8080"),
		WebOrigin:         getEnv("WEB_ORIGIN", "http://localhost:5173"),
		DatabaseURL:       mustEnv("DATABASE_URL"),
		RedisURL:          mustEnv("REDIS_URL"),
		SessionSecret:     mustEnv("SESSION_SECRET"),
		MFAEncryptionKey:  getEnv("MFA_ENCRYPTION_KEY", "00000000000000000000000000000000"),
		S3Bucket:          getEnv("S3_BUCKET", "verin-documents"),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3Endpoint:        getEnv("S3_ENDPOINT", ""),
		S3AccessKey:       getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:       getEnv("S3_SECRET_KEY", ""),
		S3UsePathStyle:    getEnv("S3_USE_PATH_STYLE", "true") == "true",
		TikaEndpoint:      getEnv("TIKA_ENDPOINT", ""),
		DefaultOrgID:      mustEnv("DEFAULT_ORG_ID"),
		SessionCookieName: "verin_session",
		CSRFCookieName:    "verin_csrf",
		SessionTTL:        24 * time.Hour,
		AuthPendingTTL:    10 * time.Minute,
		LockoutThreshold:  5,
		LockoutDuration:   15 * time.Minute,
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          getEnv("SMTP_PORT", "1025"),
		SMTPFromAddress:   getEnv("SMTP_FROM", "noreply@verin.local"),
	}

	ttlSeconds, err := strconv.Atoi(getEnv("SIGNED_URL_TTL_SECONDS", "900"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SIGNED_URL_TTL_SECONDS: %w", err)
	}
	cfg.SignedURLTTL = time.Duration(ttlSeconds) * time.Second

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("missing required env %s", key))
	}
	return value
}

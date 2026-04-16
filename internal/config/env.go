package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	defaultPort                    = "8080"
	defaultDBHost                  = "localhost"
	defaultDBPort                  = "5432"
	defaultDBUser                  = "streaming"
	defaultDBPass                  = "streaming"
	defaultDBName                  = "streaming"
	defaultDBSSL                   = "disable"
	defaultMigrateD                = "migrations"
	defaultRedisHost               = "localhost"
	defaultRedisPort               = "6379"
	defaultRedisDB                 = 0
	defaultJWTIssuer               = "demo-streaming"
	defaultJWTAccessTokenTTLSeconds int64 = 3600
	defaultJWTRefreshTokenTTLSeconds int64 = 604800
)

type SystemConfig struct {
	Port          string
	MigrationsDir string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost string
	RedisPort string
	RedisPassword string
	RedisDB       int
}

type AppConfig struct {
	JWTSecret                string
	JWTIssuer                string
	JWTAccessTokenTTLSeconds int64
	JWTRefreshTokenTTLSeconds int64
}

func LoadDotEnv() {
	_ = godotenv.Load(".env")
}

func LoadSystemConfig() SystemConfig {
	return SystemConfig{
		Port:          getEnv("PORT", defaultPort),
		MigrationsDir: getEnv("MIGRATIONS_DIR", defaultMigrateD),
		DBHost:        getEnv("DB_HOST", defaultDBHost),
		DBPort:        getEnv("DB_PORT", defaultDBPort),
		DBUser:        getEnv("DB_USER", defaultDBUser),
		DBPassword:    getEnv("DB_PASSWORD", defaultDBPass),
		DBName:        getEnv("DB_NAME", defaultDBName),
		DBSSLMode:     getEnv("DB_SSLMODE", defaultDBSSL),
		RedisHost:     getEnv("REDIS_HOST", defaultRedisHost),
		RedisPort:     getEnv("REDIS_PORT", defaultRedisPort),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       int(parseInt64("REDIS_DB", int64(defaultRedisDB))),
	}
}

func LoadAppConfig() AppConfig {
	return AppConfig{
		JWTSecret:                 getEnv("JWT_SECRET", ""),
		JWTIssuer:                 getEnv("JWT_ISSUER", defaultJWTIssuer),
		JWTAccessTokenTTLSeconds:  parsePositiveInt64("JWT_ACCESS_TOKEN_TTL_SECONDS", defaultJWTAccessTokenTTLSeconds),
		JWTRefreshTokenTTLSeconds: parsePositiveInt64("JWT_REFRESH_TOKEN_TTL_SECONDS", defaultJWTRefreshTokenTTLSeconds),
	}
}

func DatabaseURL(cfg SystemConfig) string {
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		return raw
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}

func parsePositiveInt64(key string, fallback int64) int64 {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseInt64(key string, fallback int64) int64 {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

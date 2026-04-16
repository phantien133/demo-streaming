package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultPort     = "8080"
	defaultDBHost   = "localhost"
	defaultDBPort   = "5432"
	defaultDBUser   = "streaming"
	defaultDBPass   = "streaming"
	defaultDBName   = "streaming"
	defaultDBSSL    = "disable"
	defaultMigrateD = "migrations"
)

func LoadDotEnv() {
	_ = godotenv.Load(".env")
}

func Port() string {
	return getEnv("PORT", defaultPort)
}

func MigrationsDir() string {
	return getEnv("MIGRATIONS_DIR", defaultMigrateD)
}

func DatabaseURL() string {
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		return raw
	}

	host := getEnv("DB_HOST", defaultDBHost)
	port := getEnv("DB_PORT", defaultDBPort)
	user := getEnv("DB_USER", defaultDBUser)
	pass := getEnv("DB_PASSWORD", defaultDBPass)
	name := getEnv("DB_NAME", defaultDBName)
	ssl := getEnv("DB_SSLMODE", defaultDBSSL)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	defaultPort                            = "8080"
	defaultDBHost                          = "localhost"
	defaultDBPort                          = "5432"
	defaultDBUser                          = "streaming"
	defaultDBPass                          = "streaming"
	defaultDBName                          = "streaming"
	defaultDBSSL                           = "disable"
	defaultMigrateD                        = "migrations"
	defaultRedisHost                       = "localhost"
	defaultRedisPort                       = "6379"
	defaultRedisDB                         = 0
	defaultJWTIssuer                       = "demo-streaming"
	defaultJWTAccessTokenTTLSeconds  int64 = 3600
	defaultJWTRefreshTokenTTLSeconds int64 = 604800
	// Dev/local-only defaults for media URLs.
	defaultDevSRSRTMPBaseURL           = "rtmp://localhost:1935/live"
	defaultDevSRSPlaybackOriginBaseURL = "http://localhost:8081/live"
	defaultDevSRSPlaybackCDNBaseURL    = "http://localhost:8088/live"
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

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

type AppConfig struct {
	JWTSecret                   string
	JWTIssuer                   string
	JWTAccessTokenTTLSeconds    int64
	JWTRefreshTokenTTLSeconds   int64
	DevSRSRTMPBaseURL           string
	DevSRSRTMPPublishBaseURL    string
	DevSRSPlaybackOriginBaseURL string
	DevSRSPlaybackCDNBaseURL    string

	// Internal URL for API container to fetch HLS from SRS origin (docker network).
	SRSInternalPlaybackOriginBaseURL string

	// SRS webhook (http_hooks) protection.
	SRSWebhookSecret string

	// Transcode worker config.
	TranscodeEnabled  bool
	TranscodeQueueKey string
	// TranscodeInputRTMPBaseURL when non-empty (e.g. rtmp://srs:1935/live) is used as ffmpeg input instead of HLS.
	// Avoids 404 while SRS has not written the first .m3u8 yet; ffmpeg subscribes to the same live RTMP app/stream.
	TranscodeInputRTMPBaseURL string
	TranscodeInputBaseURL     string
	// TranscodeInputWaitMaxSeconds: when using HLS input, poll until playlist returns 200 or this timeout.
	TranscodeInputWaitMaxSeconds int
	TranscodeOutputDir           string
	TranscodeFFmpegBin           string
	TranscodePreset              string
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

		// Dev/local-only media URLs. Prefer DEV_* keys; keep backwards compatibility with older SRS_* keys.
		DevSRSRTMPBaseURL:                getEnv2("DEV_SRS_RTMP_BASE_URL", "SRS_RTMP_BASE_URL", defaultDevSRSRTMPBaseURL),
		DevSRSRTMPPublishBaseURL:         getEnv2("DEV_SRS_RTMP_PUBLISH_BASE_URL", "SRS_RTMP_PUBLISH_BASE_URL", ""),
		DevSRSPlaybackOriginBaseURL:      getEnv2("DEV_SRS_PLAYBACK_ORIGIN_BASE_URL", "SRS_PLAYBACK_BASE_URL", defaultDevSRSPlaybackOriginBaseURL),
		DevSRSPlaybackCDNBaseURL:         getEnv2("DEV_SRS_PLAYBACK_CDN_BASE_URL", "SRS_PLAYBACK_CDN_BASE_URL", defaultDevSRSPlaybackCDNBaseURL),
		SRSInternalPlaybackOriginBaseURL: getEnv("SRS_INTERNAL_PLAYBACK_ORIGIN_BASE_URL", "http://srs:8080/live"),

		SRSWebhookSecret: getEnv("SRS_WEBHOOK_SECRET", ""),

		TranscodeEnabled:             parseBool("TRANSCODE_ENABLED", false),
		TranscodeQueueKey:            getEnv("TRANSCODE_QUEUE_KEY", "transcode:publish_jobs"),
		TranscodeInputRTMPBaseURL:    getEnv("TRANSCODE_INPUT_RTMP_BASE_URL", ""),
		TranscodeInputBaseURL:        getEnv2("TRANSCODE_INPUT_BASE_URL", "SRS_INTERNAL_PLAYBACK_ORIGIN_BASE_URL", "http://srs:8080/live"),
		TranscodeInputWaitMaxSeconds: int(parsePositiveInt64("TRANSCODE_INPUT_WAIT_MAX_SECONDS", 90)),
		TranscodeOutputDir:           getEnv("TRANSCODE_OUTPUT_DIR", "./tmp/transcode"),
		TranscodeFFmpegBin:           getEnv("TRANSCODE_FFMPEG_BIN", "ffmpeg"),
		TranscodePreset:              getEnv("TRANSCODE_PRESET", "veryfast"),
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

func parseBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
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

func getEnv2(primaryKey, fallbackKey, fallbackValue string) string {
	if value := os.Getenv(primaryKey); value != "" {
		return value
	}
	return getEnv(fallbackKey, fallbackValue)
}

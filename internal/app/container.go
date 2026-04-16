package app

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/cache"
	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	SystemConfig config.SystemConfig
	AppConfig    config.AppConfig

	DB         *gorm.DB
	closeDB    func() error
	Redis      *redis.Client
	closeRedis func() error
	JWTManager *auth.JWTManager
}

func NewContainer(systemCfg config.SystemConfig, appCfg config.AppConfig) (*Container, error) {
	db, closeDB, err := database.NewGormDB(config.DatabaseURL(systemCfg))
	if err != nil {
		return nil, err
	}

	redisClient, closeRedis, err := cache.NewRedisClient(systemCfg)
	if err != nil {
		_ = closeDB()
		return nil, err
	}

	jwtManager, err := auth.NewJWTManager(appCfg.JWTSecret, appCfg.JWTIssuer)
	if err != nil {
		_ = closeDB()
		_ = closeRedis()
		return nil, err
	}

	return &Container{
		SystemConfig: systemCfg,
		AppConfig:    appCfg,
		DB:           db,
		closeDB:      closeDB,
		Redis:        redisClient,
		closeRedis:   closeRedis,
		JWTManager:   jwtManager,
	}, nil
}

func (c *Container) Close() error {
	var errs []error
	if c.closeRedis != nil {
		if err := c.closeRedis(); err != nil {
			errs = append(errs, err)
		}
	}
	if c.closeDB != nil {
		if err := c.closeDB(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

package connector

import (
	"ResuMatch/internal/config"
	l "ResuMatch/pkg/logger"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

func NewRedisPool(cfg config.RedisConfig) *redis.Pool {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	return &redis.Pool{
		MaxIdle:     cfg.Pool.MaxIdle,
		MaxActive:   cfg.Pool.MaxActive,
		IdleTimeout: cfg.Pool.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address,
				redis.DialPassword(cfg.Password),
				redis.DialDatabase(cfg.DB),
				redis.DialConnectTimeout(5*time.Second),
			)
			if err != nil {
				l.Log.WithField("error", err).Error("не удалось установить соединение с Redis")
				return nil, fmt.Errorf("не удалось установить соединение с Redis: %w", err)
			}

			// проверка соединения
			if _, pingErr := conn.Do("PING"); pingErr != nil {
				closeErr := conn.Close()
				if closeErr != nil {
					return nil, fmt.Errorf("не удалось закрыть соединение с Redis: %w", closeErr)
				}
				l.Log.WithField("error", pingErr).Error("не удалось выполнить ping Redis")
				return nil, fmt.Errorf("не удалось выполнить ping Redis: %w", pingErr)
			}
			return conn, nil
		},
	}
}

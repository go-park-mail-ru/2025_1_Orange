package connector

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	l "ResuMatch/pkg/logger"
	"fmt"
<<<<<<< HEAD
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"time"
=======
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
>>>>>>> a6396a4 (Fix mistakes)
)

func NewRedisConnection(cfg config.RedisConfig) (redis.Conn, error) {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	conn, err := redis.Dial("tcp", address,
		redis.DialPassword(cfg.Password),
		redis.DialDatabase(cfg.DB),
		redis.DialConnectTimeout(5*time.Second),
	)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с Redis")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение с Redis: %w", err),
		)
	}

	if _, err := conn.Do("PING"); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping Redis")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping Redis: %w", err),
		)
	}

	return conn, nil
}

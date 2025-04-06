package redis

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	l "ResuMatch/pkg/logger"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	userSessionsPrefix = "user_sessions:"
)

type SessionRepository struct {
	conn             redis.Conn
	sessionAliveTime int
	ctx              context.Context
}

func NewSessionRepository(cfg config.RedisConfig) (repository.SessionRepository, error) {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	conn, err := redis.Dial("tcp", address,
		redis.DialPassword(cfg.Password),
		redis.DialDatabase(cfg.DB),
		redis.DialConnectTimeout(5*time.Second),
	)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось установить соединение с Redis из SessionRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить соединение с Redis: %w", err),
		)
	}

	if _, err := conn.Do("PING"); err != nil {
		l.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("не удалось выполнить ping Redis из SessionRepository")

		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось выполнить ping Redis из SessionRepository: %w", err),
		)
	}

	return &SessionRepository{
		conn:             conn,
		sessionAliveTime: cfg.TTL,
		ctx:              context.Background(),
	}, nil
}

func (r *SessionRepository) CreateSession(userID int, role string) (string, error) {
	sessionToken := uuid.NewString()

	for {
		exists, err := redis.Int(r.conn.Do("EXISTS", sessionToken))
		if err != nil {
			return "", entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить сессию для пользователя с id=%d, role=%s :%w", userID, role, err),
			)
		}
		if exists == 0 {
			break
		}
		sessionToken = uuid.NewString()
	}

	_, err := r.conn.Do("SET", sessionToken, fmt.Sprintf("%d:%s", userID, role), "EX", r.sessionAliveTime)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать сессию для пользователя с id=%d, role=%s :%w", userID, role, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = r.conn.Do("SADD", userSessionsKey, sessionToken)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось добавить сессию пользователя с id=%d, role=%s в его активные сессии :%w", userID, role, err),
		)
	}

	_, err = r.conn.Do("EXPIRE", userSessionsKey, r.sessionAliveTime)
	if err != nil {
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить TTL на сессию пользователя с id=%d, role=%s :%w", userID, role, err),
		)
	}

	return sessionToken, nil
}

func (r *SessionRepository) GetSession(sessionToken string) (int, string, error) {
	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return 0, "", entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("не удалось найти сессию с токеном=%s :%w", sessionToken, err),
			)
		}
		return 0, "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить сессию с токеном=%s :%w", sessionToken, err),
		)
	}

	var userID int
	var role string
	_, err = fmt.Sscanf(reply, "%d:%s", &userID, &role)
	if err != nil {
		return 0, "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id и role с ключом=%s :%w", reply, err),
		)
	}

	return userID, role, nil
}

func (r *SessionRepository) DeleteSession(sessionToken string) error {
	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil
		}
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить сессию с токеном=%s для удаления :%w", sessionToken, err),
		)
	}

	_, err = r.conn.Do("DEL", sessionToken)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с токеном=%s :%w", sessionToken, err),
		)
	}

	var userID int
	var role string
	_, err = fmt.Sscanf(reply, "%d:%s", &userID, &role)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id и role с ключом=%s :%w", reply, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = r.conn.Do("SREM", userSessionsKey, sessionToken)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с ключом=%s и токеном=%s из активных сессий пользователя :%w", userSessionsKey, sessionToken, err),
		)
	}

	return nil
}

func (r *SessionRepository) DeleteAllSessions(userID int, role string) error {
	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role

	sessions, err := redis.Strings(r.conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить активные сессии пользователя по ключу=%s :%w", userSessionsKey, err),
		)
	}

	for _, session := range sessions {
		_, err = r.conn.Do("DEL", session)
		if err != nil {
			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить сессию из активные сессии пользователя c ключом=%s :%w", session, err),
			)
		}
	}

	_, err = r.conn.Do("DEL", userSessionsKey)
	if err != nil {
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить ключ списка активных сессий пользователя c ключом=%s :%w", userSessionsKey, err),
		)
	}

	return nil
}

func (r *SessionRepository) Close() error {
	return r.conn.Close()
}

package redis

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	userSessionsPrefix = "user_sessions:"
)

type SessionRepository struct {
	pool             *redis.Pool
	sessionAliveTime int
	ctx              context.Context
}

func NewSessionRepository(pool *redis.Pool, ttl int) (repository.SessionRepository, error) {
	return &SessionRepository{
		pool:             pool,
		sessionAliveTime: ttl,
		ctx:              context.Background(),
	}, nil
}

func (r *SessionRepository) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        userID,
		"role":      role,
	}).Info("создание сессии в Redis CreateSession")

	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения redis: %v", err)
		}
	}()

	sessionToken := uuid.NewString()

	for {
		exists, err := redis.Int(conn.Do("EXISTS", sessionToken))
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Session Repository", "CreateSession").Inc()
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

	_, err := conn.Do("SET", sessionToken, fmt.Sprintf("%d:%s", userID, role), "EX", r.sessionAliveTime)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "CreateSession").Inc()
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать сессию для пользователя с id=%d, role=%s :%w", userID, role, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = conn.Do("SADD", userSessionsKey, sessionToken)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "CreateSession").Inc()
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось добавить сессию пользователя с id=%d, role=%s в его активные сессии :%w", userID, role, err),
		)
	}

	_, err = conn.Do("EXPIRE", userSessionsKey, r.sessionAliveTime)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "CreateSession").Inc()
		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось установить TTL на сессию пользователя с id=%d, role=%s :%w", userID, role, err),
		)
	}
	return sessionToken, nil
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionToken string) (int, string, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":    requestID,
		"sessionToken": sessionToken,
	}).Info("получение сессии в Redis GetSession")

	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения redis: %v", err)
		}
	}()

	reply, err := redis.String(conn.Do("GET", sessionToken))
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "GetSession").Inc()
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
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "GetSession").Inc()
		return 0, "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id и role с ключом=%s :%w", reply, err),
		)
	}

	return userID, role, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":    requestID,
		"sessionToken": sessionToken,
	}).Info("удаление сессии в Redis DeleteSession")

	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения redis: %v", err)
		}
	}()

	reply, err := redis.String(conn.Do("GET", sessionToken))
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteSession").Inc()
		if errors.Is(err, redis.ErrNil) {
			return nil
		}
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить сессию с токеном=%s для удаления :%w", sessionToken, err),
		)
	}

	_, err = conn.Do("DEL", sessionToken)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteSession").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с токеном=%s :%w", sessionToken, err),
		)
	}

	var userID int
	var role string
	_, err = fmt.Sscanf(reply, "%d:%s", &userID, &role)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteSession").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось распарсить сессию на id и role с ключом=%s :%w", reply, err),
		)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = conn.Do("SREM", userSessionsKey, sessionToken)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteSession").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить сессию с ключом=%s и токеном=%s из активных сессий пользователя :%w", userSessionsKey, sessionToken, err),
		)
	}

	return nil
}

func (r *SessionRepository) DeleteAllSessions(ctx context.Context, userID int, role string) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        userID,
		"role":      role,
	}).Info("удаление всех активных сессий пользователя в Redis DeleteAllSessions")

	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			l.Log.Warnf("Ошибка при закрытии соединения redis: %v", err)
		}
	}()

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role

	sessions, err := redis.Strings(conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteAllSessions").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось получить активные сессии пользователя по ключу=%s :%w", userSessionsKey, err),
		)
	}

	for _, session := range sessions {
		_, err = conn.Do("DEL", session)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteAllSessions").Inc()
			return entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить сессию из активные сессии пользователя c ключом=%s :%w", session, err),
			)
		}
	}

	_, err = conn.Do("DEL", userSessionsKey)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Session Repository", "DeleteAllSessions").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить ключ списка активных сессий пользователя c ключом=%s :%w", userSessionsKey, err),
		)
	}

	return nil
}

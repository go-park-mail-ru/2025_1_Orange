package redis

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	userSessionsPrefix = "user_sessions:"
)

type SessionDB struct {
	conn             redis.Conn
	sessionAliveTime int
	ctx              context.Context
}

func NewSessionRepository(cfg config.RedisConfig, sessionTTL int) (repository.SessionRepository, error) {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	conn, err := redis.Dial("tcp", address,
		redis.DialPassword(cfg.Password),
		redis.DialDatabase(cfg.DB),
		redis.DialConnectTimeout(5*time.Second),
	)
	if err != nil {
		return nil, entity.NewClientError(fmt.Sprintf("failed to connect to Redis: %w", err), entity.ErrRedis)
	}

	if _, err := conn.Do("PING"); err != nil {
		return nil, entity.NewClientError(fmt.Sprintf("failed to ping Redis: %w", err), entity.ErrRedis)
	}

	return &SessionDB{
		conn:             conn,
		sessionAliveTime: sessionTTL,
		ctx:              context.Background(),
	}, nil
}

func (r *SessionDB) CreateSession(userID int, role string) (string, error) {
	sessionToken := uuid.NewString()

	for {
		exists, err := redis.Int(r.conn.Do("EXISTS", sessionToken))
		if err != nil {
			return "", entity.NewClientError(fmt.Sprintf("failed to check session existence: %w", err), entity.ErrRedis)
		}
		if exists == 0 {
			break
		}
		sessionToken = uuid.NewString()
	}

	_, err := r.conn.Do("SET", sessionToken, fmt.Sprintf("%d:%s", userID, role), "EX", r.sessionAliveTime)
	if err != nil {
		return "", entity.NewClientError(fmt.Sprintf("failed to create session: %w", err), entity.ErrRedis)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = r.conn.Do("SADD", userSessionsKey, sessionToken)
	if err != nil {
		return "", entity.NewClientError(fmt.Sprintf("failed to add session to user sessions: %w", err), entity.ErrRedis)
	}

	_, err = r.conn.Do("EXPIRE", userSessionsKey, r.sessionAliveTime)
	if err != nil {
		return "", entity.NewClientError(fmt.Sprintf("failed to set TTL for user sessions: %w", err), entity.ErrRedis)
	}

	return sessionToken, nil
}

func (r *SessionDB) GetSession(sessionToken string) (int, string, error) {
	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return 0, "", entity.NewClientError(fmt.Sprintf("session with token: %s not found", sessionToken), entity.ErrNotFound)
		}
		return 0, "", entity.NewClientError(fmt.Sprintf("failed to get session: %w", err), entity.ErrRedis)
	}

	var userID int
	var role string
	_, err = fmt.Sscanf(reply, "%d:%s", &userID, &role)
	if err != nil {
		return 0, "", entity.NewClientError(fmt.Sprintf("failed to parse session data: %w", err), entity.ErrRedis)
	}

	return userID, role, nil
}

func (r *SessionDB) DeleteSession(sessionToken string) error {
	reply, err := redis.String(r.conn.Do("GET", sessionToken))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil
		}
		return entity.NewClientError(fmt.Sprintf("failed to get session for deletion: %w", err), entity.ErrRedis)
	}

	_, err = r.conn.Do("DEL", sessionToken)
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to delete session for deletion: %w", err), entity.ErrRedis)
	}

	var userID int
	var role string
	_, err = fmt.Sscanf(reply, "%d:%s", &userID, &role)
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to parse session data: %w", err), entity.ErrRedis)
	}

	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role
	_, err = r.conn.Do("SREM", userSessionsKey, sessionToken)
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to remove session from user sessions: %w", err), entity.ErrRedis)
	}

	return nil
}

func (r *SessionDB) DeleteAllSessions(userID int, role string) error {
	userSessionsKey := userSessionsPrefix + strconv.Itoa(userID) + ":" + role

	sessions, err := redis.Strings(r.conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to get sessions for removal: %w", err), entity.ErrRedis)
	}

	for _, session := range sessions {
		_, err = r.conn.Do("DEL", session)
		if err != nil {
			return entity.NewClientError(fmt.Sprintf("failed to remove session from user sessions: %w", err), entity.ErrRedis)
		}
	}

	_, err = r.conn.Do("DEL", userSessionsKey)
	if err != nil {
		return entity.NewClientError(fmt.Sprintf("failed to delete user sessions list: %w", err), entity.ErrRedis)
	}

	return nil
}

func (r *SessionDB) Close() error {
	return r.conn.Close()
}

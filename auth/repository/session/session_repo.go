package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var mutex sync.RWMutex

type DbRedisCfg struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	DbNumber int    `yaml:"db"`
	Timer    int    `yaml:"timer"`
}

type SessionRepo struct {
	sessionRedisClient *redis.Client
	Connection         bool
}

func (redisRepo *SessionRepo) CheckRedisSessionConnection(sessionCfg DbRedisCfg) {
	ctx := context.Background()
	for {
		_, err := redisRepo.sessionRedisClient.Ping(ctx).Result()
		mutex.Lock()
		mutex.RLock()
		redisRepo.Connection = err == nil
		mutex.Unlock()
		mutex.RUnlock()
		time.Sleep(time.Duration(sessionCfg.Timer) * time.Second)
	}
}

func GetSessionRepo(sessionCfg DbRedisCfg) (*SessionRepo, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     sessionCfg.Host,
		Password: sessionCfg.Password,
		DB:       sessionCfg.DbNumber,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	sessionRepo := SessionRepo{
		sessionRedisClient: redisClient,
		Connection:         true,
	}

	go sessionRepo.CheckRedisSessionConnection(sessionCfg)

	return &sessionRepo, nil
}

func (redisRepo *SessionRepo) AddSession(ctx context.Context, active Session) (bool, error) {
	if !redisRepo.Connection {
		fmt.Printf("Redis session connection lost")
		return false, nil
	}

	redisRepo.sessionRedisClient.Set(ctx, active.SID, active.Login, 24*time.Hour)

	sessionAdded, err_check := redisRepo.CheckActiveSession(ctx, active.SID)

	if err_check != nil {
		return false, err_check
	}

	return sessionAdded, nil
}

func (redisRepo *SessionRepo) GetUserLogin(ctx context.Context, sid string) (string, error) {
	if !redisRepo.Connection {
		fmt.Printf("Redis session connection lost")
		return "", nil
	}

	value, err := redisRepo.sessionRedisClient.Get(ctx, sid).Result()
	if err != nil {
		fmt.Printf("Error, cannot find session " + sid)
		return "", err
	}

	return value, nil
}

func (redisRepo *SessionRepo) CheckActiveSession(ctx context.Context, sid string) (bool, error) {
	if !redisRepo.Connection {
		fmt.Printf("Redis session connection lost")
		return false, nil
	}

	_, err := redisRepo.sessionRedisClient.Get(ctx, sid).Result()
	if err == redis.Nil {
		fmt.Printf("Key " + sid + " not found")
		return false, nil
	}

	if err != nil {
		fmt.Printf("Get request could not be completed ", err)
		return false, err
	}

	return true, nil
}

func (redisRepo *SessionRepo) DeleteSession(ctx context.Context, sid string) (bool, error) {
	_, err := redisRepo.sessionRedisClient.Del(ctx, sid).Result()
	if err != nil {
		fmt.Errorf("Delete request could not be completed:", err)
		return false, err
	}

	return true, nil
}

package redis

import (
	"ResuMatch/internal/domain/mocks"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type ISessionRepository interface {
	CreateSession(session *mocks.Session) error
	FindSession(sid string) (*mocks.Session, error)
	DeleteSession(sessID string) error
}

type SessionRedisRepository struct {
	db *redis.Pool
}

func NewSessionRedisRepository(db *redis.Pool) *SessionRedisRepository {
	return &SessionRedisRepository{
		db: db,
	}
}

func (s *SessionRedisRepository) CreateSession(session *mocks.Session) error {
	conn := s.db.Get()
	defer conn.Close()
	dataSerialized, err := json.Marshal(session)
	if err != nil {
		return err
	}
	mkey := "sessions:" + session.SID

	res, err := redis.String(conn.Do("SET", mkey, dataSerialized, "EX", 86400))
	if err != nil {
		return err
	}

	if res != "OK" {
		return fmt.Errorf("Not OK")
	}

	return nil
}

func (s *SessionRedisRepository) FindSession(sid string) (*mocks.Session, error) {
	conn := s.db.Get()
	defer conn.Close()
	mkey := "sessions:" + sid
	data, err := redis.String(conn.Do("GET", mkey))
	if err != nil {
		return nil, err
	}

	sess := &mocks.Session{}
	err = json.Unmarshal([]byte(data), sess)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *SessionRedisRepository) DeleteSession(sessID string) error {
	conn := s.db.Get()
	defer conn.Close()
	mkey := "sessions:" + sessID
	_, err := redis.Int(conn.Do("DEL", mkey))
	if err != nil {
		return err
	}
	return nil
}

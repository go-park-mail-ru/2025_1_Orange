package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type SessionStorage struct {
	Sessions map[string]uint64
}

func NewSessionStorage() *SessionStorage {
	return &SessionStorage{
		Sessions: make(map[string]uint64),
	}
}
func CreateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("createSessionID: failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *SessionStorage) createSession(_ context.Context, userID uint64, sid string) error {
	s.Sessions[sid] = userID
	return nil
}

func (s *SessionStorage) GetSession(sid string) (uint64, error) {
	userID, exists := s.Sessions[sid]
	if !exists {
		return 0, fmt.Errorf("session not found")
	}
	return userID, nil
}

func (s *SessionStorage) DeleteSession(sid string) error {
	delete(s.Sessions, sid)
	return nil
}

func (s *SessionStorage) CreateSession(ctx context.Context, userID uint64) (string, error) {
	sid, err := CreateSessionID()
	if err != nil {
		return "", fmt.Errorf("CreateSession: can't generate session ID for user %d: %w", userID, err)
	}

	err = s.createSession(ctx, userID, sid)
	if err != nil {
		return "", fmt.Errorf("CreateSession: can't create session for user %d with sid %s: %w", userID, sid, err)
	}
	return sid, nil
}
func (s *SessionStorage) FindActiveSession(_ context.Context, sid string) (uint64, error) {
	userID, err := s.GetSession(sid)
	if err != nil {
		return 0, fmt.Errorf("FindActiveSession: can't get session %s: %w", sid, err)
	}
	return userID, nil
}

func (s *SessionStorage) KillSession(_ context.Context, sid string) error {
	err := s.DeleteSession(sid)
	if err != nil {
		return fmt.Errorf("KillSession: can't delete session %s: %w", sid, err)
	}
	return nil
}

func (s *SessionStorage) GetUserIDFromSession(sid string) (uint64, error) {
	userID, err := s.GetSession(sid)
	if err != nil {
		return 0, fmt.Errorf("GetUserIDFromSession: can't get session %s: %w", sid, err)
	}

	return userID, nil
}

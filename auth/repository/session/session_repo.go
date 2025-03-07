package session

import (
	"context"
	"fmt"
)

var Sessions = make(map[string]uint64)

type StaticSessionRepo struct{}

func (r *StaticSessionRepo) CreateSession(ctx context.Context, userID uint64, sid string) error {
	Sessions[sid] = userID
	return nil
}

func (r *StaticSessionRepo) GetSession(sessionID string) (uint64, error) {
	userID, ok := Sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found")
	}
	return userID, nil
}
func (r *StaticSessionRepo) DeleteSession(sessionID string) error {
	delete(Sessions, sessionID)
	return nil
}

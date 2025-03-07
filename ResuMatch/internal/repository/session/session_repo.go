package session

import (
	"context"
	"fmt"
)

var Sessions = make(map[string]uint64)

type Sessionrepo struct{}

func (r *Sessionrepo) CreateSession(_ context.Context, userID uint64, sid string) error {
	Sessions[sid] = userID
	return nil
}

func (r *Sessionrepo) GetSession(sessionID string) (uint64, error) {
	userID, ok := Sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found")
	}
	return userID, nil
}
func (r *Sessionrepo) DeleteSession(sessionID string) error {
	delete(Sessions, sessionID)
	return nil
}

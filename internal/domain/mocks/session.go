package mocks

import "time"

type Session struct {
	UserID  int
	SID     string
	Expires time.Time
}

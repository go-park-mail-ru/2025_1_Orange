package session

import "time"

type Session struct {
	Login   string
	SID     string
	Expires time.Time
}

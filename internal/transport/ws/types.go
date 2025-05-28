package ws

import (
	"ResuMatch/internal/entity"
	"time"
)

type MessageType string

const (
	MessageTypeChat         MessageType = "message"
	MessageTypeNotification MessageType = "notification"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 55 * time.Second
	writeWait  = 10 * time.Second
)

type ConnectionKey struct {
	UserID int
	Type   entity.UserRole
}

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

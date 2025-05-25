package ws

import "ResuMatch/internal/entity"

type MessageType string

const (
	MessageTypeChat         MessageType = "message"
	MessageTypeNotification MessageType = "notification"
)

type ConnectionKey struct {
	UserID int
	Type   entity.UserRole
}

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

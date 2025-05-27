package dto

import (
	"ResuMatch/internal/entity"
	"time"
)

// easyjson:json
type MessageResponse struct {
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	SenderID      int       `json:"sender_id"`
	ReceiverID    int       `json:"receiver_id"`
	Avatar        string    `json:"avatar"`
	FromApplicant bool      `json:"from_applicant"`
	Payload       string    `json:"payload"`
	SentAt        time.Time `json:"sent_at"`
}

// easyjson:json
type MessagesResponseList []*MessageResponse

// easyjson:json
type MessageRequest struct {
	ChatID     int             `json:"chat_id"`
	SenderID   int             `json:"sender_id"`
	ReceiverID int             `json:"receiver_id"`
	SenderRole entity.UserRole `json:"sender_role"`
	Payload    string          `json:"payload"`
}

package dto

import "time"

type MessageResponse struct {
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	SenderID      int       `json:"sender_id"`
	Avatar        string    `json:"avatar"`
	FromApplicant bool      `json:"from_applicant"`
	Payload       string    `json:"payload"`
	SentAt        time.Time `json:"sent_at"`
}

package entity

import "time"

type Message struct {
	ID            int       `json:"id"`
	ChatID        int       `json:"chat_id"`
	SenderID      int       `json:"sender_id"`
	FromApplicant bool      `json:"from_applicant"`
	Payload       string    `json:"payload"`
	SentAt        time.Time `json:"sent_at"`
}

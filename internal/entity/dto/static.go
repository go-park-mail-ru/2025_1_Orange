package dto

import "time"

type UploadStaticResponse struct {
	ID        int       `json:"id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

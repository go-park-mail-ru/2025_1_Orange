package entity

import "time"

type Static struct {
	ID        int       `json:"id"`
	FilePath  string    `json:"file_path"`
	FileName  string    `json:"file_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

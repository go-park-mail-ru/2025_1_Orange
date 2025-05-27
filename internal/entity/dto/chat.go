package dto

import "time"

// easyjson:json
type ChatResponse struct {
	ID        int                  `json:"id"`
	Vacancy   *VacancyChatResponse `json:"vacancy"`
	Resume    *ResumeChatResponse  `json:"resume"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// easyjson:json
type ChatShortResponse struct {
	ID           int             `json:"id"`
	VacancyTitle string          `json:"vacancy_title"`
	User         ChatUserPreview `json:"user"`
}

// easyjson:json
type ChatUserPreview struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	AvatarPath string `json:"avatar_path"`
}

// easyjson:json
type ChatResponseList []*ChatShortResponse

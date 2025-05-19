package dto

import "time"

type ChatResponse struct {
	ID        int                  `json:"id"`
	Vacancy   *VacancyChatResponse `json:"vacancy"`
	Resume    *ResumeChatResponse  `json:"resume"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

type ApplicantChatResponse struct {
	ID           int                        `json:"id"`
	Employer     *ChatShortResponseEmployer `json:"employer"`
	VacancyTitle string                     `json:"vacancy_title"`
}

type EmployerChatResponse struct {
	ID           int                         `json:"id"`
	Applicant    *ChatShortResponseApplicant `json:"applicant"`
	VacancyTitle string                      `json:"vacancy_title"`
}

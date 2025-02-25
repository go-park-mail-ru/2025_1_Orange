package models

type Vacancy struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Salary      string `json:"salary"`
	Description string `json:"description"`
}

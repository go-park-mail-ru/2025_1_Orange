package models

type Vacancy struct {
	ID         int     `json:"id"`
	Profession string  `json:"profession"`
	Salary     string  `json:"salary"`
	Company    string  `json:"company"`
	City       string  `json:"city"`
	Badges     []Badge `json:"badges"`
	DayCreated int     `json:"day_created"`
	Count      int     `json:"count"`
}

type Badge struct {
	Name string `json:"name"`
}

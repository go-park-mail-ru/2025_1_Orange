// package models

// import "time"

// type Vacancy struct {
// 	ID          int       `json:"id"`
// 	Title       string    `json:"title"`
// 	Company     string    `json:"company"`
// 	Location    string    `json:"location"`
// 	Salary      string    `json:"salary"`
// 	Description string    `json:"description"`
// 	CreatedAt   time.Time `json:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at"`
// 	Active      bool      `json:"active"`
// 	PostedBy    int       `json:"posted_by"`   // ID пользователя, разместившего вакансию
// 	EmployerID  int       `json:"employer_id"` // ID компании-работодателя
// }

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

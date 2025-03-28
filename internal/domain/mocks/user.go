package mocks

import "time"

type User struct {
	ID             uint64 `json:"id"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	CompanyName    string `json:"companyName"`
	CompanyAddress string `json:"companyAddress"`
}

type DBUser struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	MiddleName        string `json:"middleName"`
	City              string `json:"city"`
	Birthday          string `json:"birthdate"`
	Gender            string `json:"gender"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	JobSearchStatusId uint64 `json:"jobSearchStatusId"`
	ProfileImageId    uint64 `json:"profileImageId"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UserUpdateInfo struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	MiddleName string `json:"middleName"`
	City       string `json:"city"`
	Birthday   string `json:"birthdate"`
	Gender     string `json:"gender"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

//{id} -> {first_name, last_name, middle_name, city, birth_date, gender, email, password, job_search_status_id, profile_image_id, created_at, updated_at}`

package data

import (
	"ResuMatch/internal/models"
)

var Users = map[string]models.User{
	"user1": {
		ID:             1,
		Email:          "john.doe@example.com",
		Password:       "qwertyuiop",
		FirstName:      "John",
		LastName:       "Doe",
		CompanyName:    "Acme Corporation",
		CompanyAddress: "123 Main Street, Anytown, CA 91234",
	},
	"user2": {
		ID:             2,
		Email:          "jane.smith@gmail.com",
		Password:       "asdfghjkl",
		FirstName:      "Jane",
		LastName:       "Smith",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user3": {
		ID:             3,
		Email:          "anonymous@example.net",
		Password:       "zxcvbnm",
		FirstName:      "",
		LastName:       "",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user4": {
		ID:             4,
		Email:          "special.user@domain.org",
		Password:       "qqqqqqqqqq",
		FirstName:      "Special",
		LastName:       "User",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user5 ": {
		ID:             5,
		Email:          "a.very.long.email.address.for.testing@very.long.domain.example.com",
		Password:       "1234567890",
		FirstName:      "A Very Long First Name For Testing Purposes",
		LastName:       "An Equally Long Last Name Also For Testing Purposes",
		CompanyName:    "A Company With A Very Long Name To See How It Handles",
		CompanyAddress: "A Very Long Address That Spans Multiple Lines And Contains All Sorts Of Characters And Punctuation, 456 Elm Street, Suite 789, Somecity, Somestate, 90210",
	},
}

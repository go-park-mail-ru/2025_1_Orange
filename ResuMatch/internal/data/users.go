package data

import (
	"ResuMatch/internal/models"
)

var Users = map[string]models.User{
	"user1": {
		ID:             1,
		Email:          "john.doe@example.com",
		Password:       "$2a$10$XXXXXXXXXXXXXX",
		FirstName:      "John",
		LastName:       "Doe",
		CompanyName:    "Acme Corporation",
		CompanyAddress: "123 Main Street, Anytown, CA 91234",
	},
	"user2": {
		ID:             2,
		Email:          "jane.smith@gmail.com",
		Password:       "$2a$10$YYYYYYYYYYYYYY",
		FirstName:      "Jane",
		LastName:       "Smith",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user3": {
		ID:             3,
		Email:          "anonymous@example.net",
		Password:       "$2a$10$zzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		FirstName:      "",
		LastName:       "",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user4": {
		ID:             4,
		Email:          "special.user@domain.org",
		Password:       "$2a$10$qqqqqqqqqqqqqqqqqqqqqqqqqqqq",
		FirstName:      "Special",
		LastName:       "User",
		CompanyName:    "",
		CompanyAddress: "",
	},
	"user5 ": {
		ID:             5,
		Email:          "a.very.long.email.address.for.testing@very.long.domain.example.com",
		Password:       "$2a$10$rrrrrrrrrrrrrrrrrrrrrrrrrrrrr",
		FirstName:      "A Very Long First Name For Testing Purposes",
		LastName:       "An Equally Long Last Name Also For Testing Purposes",
		CompanyName:    "A Company With A Very Long Name To See How It Handles",
		CompanyAddress: "A Very Long Address That Spans Multiple Lines And Contains All Sorts Of Characters And Punctuation, 456 Elm Street, Suite 789, Somecity, Somestate, 90210",
	},
}

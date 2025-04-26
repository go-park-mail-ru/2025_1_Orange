package entity

type Vote struct {
	ID     int
	PollID int
	UserID int
	Role   string
	Answer int
}

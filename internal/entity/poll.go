package entity

type Vote struct {
	ID     int
	PollID int
	UserID int
	Role   string
	Answer int
}

type Poll struct {
	ID   int
	Name string
}

type VoteStats struct {
	Answer int
	Count  int
}

package dto

type VotePollRequest struct {
	PollID int `json:"poll_id" valid:"required"`
	Answer int `json:"answer" valid:"required,range(1|5)"`
}

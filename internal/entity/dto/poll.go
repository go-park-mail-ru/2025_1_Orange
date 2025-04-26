package dto

type VotePollRequest struct {
	PollID int `json:"poll_id" valid:"required"`
	Answer int `json:"answer" valid:"required,range(1|5)"`
}

type PollStatsResponse struct {
	PollID  int          `json:"poll_id"`
	Average float64      `json:"average"`
	Name    string       `json:"name"`
	Stars   []*StarStats `json:"stars"`
}

type StarStats struct {
	Star       int     `json:"star"`
	Amount     int     `json:"amount"`
	Percentage float64 `json:"percentage"`
}

package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"net/http"
)

type PollHandler struct {
	auth usecase.Auth
	poll usecase.Poll
}

func NewPollHandler(auth usecase.Auth, poll usecase.Poll) *PollHandler {
	return &PollHandler{auth: auth, poll: poll}
}

func (h *PollHandler) Configure(r *http.ServeMux) {
	pollMux := http.NewServeMux()

	pollMux.HandleFunc("POST /vote", h.Vote)

	r.Handle("/poll/", http.StripPrefix("/poll", pollMux))
}

func (h *PollHandler) Vote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	var voteDTO dto.VotePollRequest
	if err := json.NewDecoder(r.Body).Decode(&voteDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// TODO
	err = h.poll.Vote(ctx, userID, role, &voteDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

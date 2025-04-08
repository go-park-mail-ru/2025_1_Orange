package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
)

type VacancyHandler struct {
	auth    usecase.Auth
	vacancy usecase.Vacancy
	cfg     config.CSRFConfig
}

func NewVacancyHandler(auth usecase.Auth, vac usecase.Vacancy, cfg config.CSRFConfig) VacancyHandler {
	return VacancyHandler{
		auth:    auth,
		vacancy: vac,
		cfg:     cfg,
	}
}

func (h *VacancyHandler) Configure(r *http.ServeMux) {
	vacancyMux := http.NewServeMux()

	vacancyMux.HandleFunc("POST /", h.CreateVacancy)
	vacancyMux.HandleFunc("GET /{id}", h.GetVacancy)
	vacancyMux.HandleFunc("PUT /{id}", h.UpdateVacancy)
	vacancyMux.HandleFunc("DELETE /{id}", h.DeleteVacancy)
	vacancyMux.HandleFunc("GET /employer/{employer_id}", h.GetByEmployerID)

	r.Handle("/vacancy/", http.StripPrefix("/vacancy", vacancyMux))
}

func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	var vacancyDTO entity.Vacancy
	if err := json.NewDecoder(r.Body).Decode(&vacancyDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancy, err := h.vacancy.CreateVacancy(ctx, &dto.UserFromSession{
		ID:       currentUserID,
		UserType: userType,
	})
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(vacancy)
}

func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancy, err := h.vacancy.GetVacancy(ctx, int(id))
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(vacancy)
}

func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	var vacancyDTO entity.Vacancy
	if err := json.NewDecoder(r.Body).Decode(&vacancyDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancy, err := h.vacancy.UpdateVacancy(ctx, int(id), &vacancyDTO, &dto.UserFromSession{
		ID:       currentUserID,
		UserType: userType,
	})
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(vacancy)
}

func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	err = h.vacancy.DeleteVacancy(ctx, int(id), &dto.UserFromSession{
		ID:       currentUserID,
		UserType: userType,
	})
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *VacancyHandler) GetByEmployerID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	employerIDStr := r.PathValue("employer_id")
	employerID, err := strconv.ParseInt(employerIDStr, 10, 64)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancies, err := h.vacancy.GetVacanciesByEmpID(ctx, int(employerID))
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(vacancies)
}

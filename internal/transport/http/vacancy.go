package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"context"
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
	vacancyMux.HandleFunc("GET /vacancies", h.GetAllVacancies)
	vacancyMux.HandleFunc("POST /vacancies", h.CreateVacancy)
	vacancyMux.HandleFunc("GET /vacancy/{id}", h.GetVacancy)
	vacancyMux.HandleFunc("PUT /vacancy/{id}", h.UpdateVacancy)
	vacancyMux.HandleFunc("DELETE /vacancy/{id}", h.DeleteVacancy)
	vacancyMux.HandleFunc("POST /vacancy/{id}/response", h.ApplyToVacancy)
	r.Handle("/vacancy/", http.StripPrefix("/vacancy", vacancyMux))
}

func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	var vacancyCreate dto.VacancyCreate
	if err := json.NewDecoder(r.Body).Decode(&vacancyCreate); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	ctx = context.WithValue(ctx, "employerID", currentUserID)

	vacancy, err := h.vacancy.CreateVacancy(ctx, &vacancyCreate)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancy, err := h.vacancy.GetVacancy(ctx, vacancyID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	var vacancyUpdate dto.VacancyUpdate
	if err := json.NewDecoder(r.Body).Decode(&vacancyUpdate); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	ctx = context.WithValue(ctx, "employerID", currentUserID)

	vacancy, err := h.vacancy.UpdateVacancy(ctx, vacancyID, &vacancyUpdate)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
	return

}

func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	response, err := h.vacancy.DeleteVacancy(ctx, id, currentUserID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetAllVacancies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	// _, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	// if err != nil {
	// 	utils.WriteAPIError(w, utils.ToAPIError(err))
	// 	return
	// }

	// if role != "applicant" {
	// 	utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
	// 	return
	// }

	vacancies, err := h.vacancy.GetAll(ctx)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
func (h *VacancyHandler) ApplyToVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}
	// Получаем ID вакансии из URL
	vacancyID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Получаем ID текущего пользователя из контекста аутентификации
	applicantID, _, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}
	var req dto.ApplyToVacancyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	err = h.vacancy.ApplyToVacancy(ctx, vacancyID, applicantID, req.ResumeID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

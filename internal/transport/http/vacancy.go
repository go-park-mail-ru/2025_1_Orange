package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
<<<<<<< HEAD
	"context"
=======
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
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
<<<<<<< HEAD
	vacancyMux.HandleFunc("GET /vacancies", h.GetAllVacancies)
	vacancyMux.HandleFunc("POST /vacancies", h.CreateVacancy)
	vacancyMux.HandleFunc("GET /vacancy/{id}", h.GetVacancy)
	vacancyMux.HandleFunc("PUT /vacancy/{id}", h.UpdateVacancy)
	vacancyMux.HandleFunc("DELETE /vacancy/{id}", h.DeleteVacancy)
	vacancyMux.HandleFunc("POST /vacancy/{id}/response", h.ApplyToVacancy)
=======

	vacancyMux.HandleFunc("POST /", h.CreateVacancy)
	vacancyMux.HandleFunc("GET /{id}", h.GetVacancy)
	vacancyMux.HandleFunc("PUT /{id}", h.UpdateVacancy)
	vacancyMux.HandleFunc("DELETE /{id}", h.DeleteVacancy)
	vacancyMux.HandleFunc("GET /employer/{employer_id}", h.GetByEmployerID)

>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	r.Handle("/vacancy/", http.StripPrefix("/vacancy", vacancyMux))
}

func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

<<<<<<< HEAD
=======
	// Проверка сессии
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

<<<<<<< HEAD
	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
=======
	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

<<<<<<< HEAD
	var vacancyCreate dto.VacancyCreate
	if err := json.NewDecoder(r.Body).Decode(&vacancyCreate); err != nil {
=======
	var vacancyDTO entity.Vacancy
	if err := json.NewDecoder(r.Body).Decode(&vacancyDTO); err != nil {
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

<<<<<<< HEAD
	ctx = context.WithValue(ctx, "employerID", currentUserID)

	vacancy, err := h.vacancy.CreateVacancy(ctx, &vacancyCreate)
=======
	vacancy, err := h.vacancy.CreateVacancy(ctx, &dto.UserFromSession{
		ID:       currentUserID,
		UserType: userType,
	})
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
<<<<<<< HEAD
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
=======
	_ = json.NewEncoder(w).Encode(vacancy)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}

func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
<<<<<<< HEAD
	vacancyID, err := strconv.Atoi(idStr)
=======
	id, err := strconv.ParseInt(idStr, 10, 64)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

<<<<<<< HEAD
	vacancy, err := h.vacancy.GetVacancy(ctx, vacancyID)
=======
	vacancy, err := h.vacancy.GetVacancy(ctx, int(id))
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
<<<<<<< HEAD
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
=======
	_ = json.NewEncoder(w).Encode(vacancy)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}

func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

<<<<<<< HEAD
	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
=======
	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

<<<<<<< HEAD
	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
=======
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

<<<<<<< HEAD
	var vacancyUpdate dto.VacancyUpdate
	if err := json.NewDecoder(r.Body).Decode(&vacancyUpdate); err != nil {
=======
	var vacancyDTO entity.Vacancy
	if err := json.NewDecoder(r.Body).Decode(&vacancyDTO); err != nil {
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

<<<<<<< HEAD
	ctx = context.WithValue(ctx, "employerID", currentUserID)

	vacancy, err := h.vacancy.UpdateVacancy(ctx, vacancyID, &vacancyUpdate)
=======
	vacancy, err := h.vacancy.UpdateVacancy(ctx, int(id), &vacancyDTO, &dto.UserFromSession{
		ID:       currentUserID,
		UserType: userType,
	})
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
<<<<<<< HEAD
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
=======
	_ = json.NewEncoder(w).Encode(vacancy)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}

func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

<<<<<<< HEAD
	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
=======
	currentUserID, userType, err := h.auth.GetUserIDBySession(cookie.Value)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

<<<<<<< HEAD
	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
=======
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

<<<<<<< HEAD
	response, err := h.vacancy.DeleteVacancy(ctx, id, currentUserID)
=======
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
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
<<<<<<< HEAD
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

	_, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if role != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

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

	var req dto.VacancyResponse
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
=======
	_ = json.NewEncoder(w).Encode(vacancies)
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
}

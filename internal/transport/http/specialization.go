package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"net/http"
)

type SpecializationHandler struct {
	specialization usecase.SpecializationUsecase
	// cfg            config.CSRFConfig
}

// NewSpecializationHandler создает новый экземпляр SpecializationHandler
func NewSpecializationHandler(specialization usecase.SpecializationUsecase) SpecializationHandler {
	return SpecializationHandler{specialization: specialization}
}

func (h *SpecializationHandler) Configure(r *http.ServeMux) {
	specializationMux := http.NewServeMux()
	specializationMux.HandleFunc("GET /all", h.GetAllSpecializationNames)

	r.Handle("/specialization/", http.StripPrefix("/specialization", specializationMux))
}

// GetAllSpecializationNames godoc
// @Tags Specialization
// @Summary Получение списка всех специализаций
// @Description Возвращает список имен всех специализаций без ID
// @Produce json
// @Success 200 {object} dto.SpecializationNamesResponse "Список имен специализаций"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /specialization/all [get]
func (h *SpecializationHandler) GetAllSpecializationNames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем список имен специализаций
	specializationNames, err := h.specialization.GetAllSpecializationNames(ctx)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(specializationNames); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

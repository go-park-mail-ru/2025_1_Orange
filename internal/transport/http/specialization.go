package http

import (
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
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
	specializationMux.HandleFunc("GET /salaries", h.GetSpecializationSalaries)

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
	if err := utils.WriteJSON(w, specializationNames); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

// GetSpecializationSalaries godoc
// @Tags Specialization
// @Summary Получение вилок зарплат по специализациям
// @Description Возвращает минимальную, максимальную и среднюю зарплату для каждой специализации
// @Produce json
// @Success 200 {object} dto.SpecializationSalaryRangesResponse "Вилки зарплат по специализациям"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /specialization/salaries [get]
func (h *SpecializationHandler) GetSpecializationSalaries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем данные о зарплатах
	salaryRanges, err := h.specialization.GetSpecializationSalaries(ctx)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	if err := utils.WriteJSON(w, salaryRanges); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

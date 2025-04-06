package utils

import (
	"ResuMatch/internal/entity"
	"encoding/json"
	"errors"
	"net/http"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func ToAPIError(err error) APIError {
	var apiError APIError
	var customError entity.Error
	if errors.As(err, &customError) {
		apiError.Message = customError.AppErr().Error()
		svcError := customError.SvcErr()
		switch {
		case errors.Is(svcError, entity.ErrNotFound):
			apiError.Status = http.StatusNotFound
		case errors.Is(svcError, entity.ErrBadRequest):
			apiError.Status = http.StatusBadRequest
		case errors.Is(svcError, entity.ErrUnauthorized):
			apiError.Status = http.StatusUnauthorized
		case errors.Is(svcError, entity.ErrForbidden):
			apiError.Status = http.StatusForbidden
		case errors.Is(svcError, entity.ErrAlreadyExists):
			apiError.Status = http.StatusConflict
		case errors.Is(svcError, entity.ErrInternal):
			apiError.Status = http.StatusInternalServerError
		default:
			apiError.Status = http.StatusInternalServerError
		}
	}
	return apiError
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteAPIError(w, APIError{
		Status:  status,
		Message: err.Error(),
	})
}

func WriteAPIError(w http.ResponseWriter, apiError APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)
	err := json.NewEncoder(w).Encode(apiError)
	if err != nil {
		return
	}
}

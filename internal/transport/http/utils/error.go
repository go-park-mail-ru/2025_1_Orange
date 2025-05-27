package utils

import (
	"ResuMatch/internal/entity"
	l "ResuMatch/pkg/logger"
	"encoding/json"
	"errors"
	"net/http"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var errorToStatus = map[error]int{
	entity.ErrNotFound:      http.StatusNotFound,
	entity.ErrBadRequest:    http.StatusBadRequest,
	entity.ErrUnauthorized:  http.StatusUnauthorized,
	entity.ErrForbidden:     http.StatusForbidden,
	entity.ErrAlreadyExists: http.StatusConflict,
	entity.ErrInternal:      http.StatusInternalServerError,
}

func ToAPIError(err error) APIError {
	var apiError APIError
	var customError entity.Error
	if errors.As(err, &customError) {
		apiError.Message = customError.InternalErr().Error()
		svcError := customError.ClientErr()
		if status, found := errorToStatus[svcError]; found {
			apiError.Status = status
		} else {
			apiError.Status = http.StatusInternalServerError
		}
	} else {
		apiError.Message = "internal server error"
		apiError.Status = http.StatusInternalServerError
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
	if apiError.Status < 500 {
		l.Log.Warnf("StatusCode: %d, Message: %s", apiError.Status, apiError.Message)
	} else {
		l.Log.Errorf("StatusCode: %d, Message: %s", apiError.Status, apiError.Message)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)
	err := json.NewEncoder(w).Encode(apiError)
	if err != nil {
		return
	}
}

package utils

import (
	"ResuMatch/internal/entity"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToAPIError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		inputError     error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Custom error - Not Found",
			inputError:     entity.NewError(entity.ErrNotFound, errors.New("not found")),
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "not found",
		},
		{
			name:           "Custom error - Bad Request",
			inputError:     entity.NewError(entity.ErrBadRequest, errors.New("bad request")),
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "bad request",
		},
		{
			name:           "Custom error - Unauthorized",
			inputError:     entity.NewError(entity.ErrUnauthorized, errors.New("unauthorized")),
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "unauthorized",
		},
		{
			name:           "Custom error - Forbidden",
			inputError:     entity.NewError(entity.ErrForbidden, errors.New("forbidden")),
			expectedStatus: http.StatusForbidden,
			expectedMsg:    "forbidden",
		},
		{
			name:           "Custom error - Conflict",
			inputError:     entity.NewError(entity.ErrAlreadyExists, errors.New("conflict")),
			expectedStatus: http.StatusConflict,
			expectedMsg:    "conflict",
		},
		{
			name:           "Custom error - Internal",
			inputError:     entity.NewError(entity.ErrInternal, errors.New("internal error")),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal error",
		},
		{
			name:           "Standard error",
			inputError:     errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal server error",
		},
		{
			name:           "Custom error with unregistered client error",
			inputError:     entity.NewError(errors.New("unknown"), errors.New("some error")),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "some error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ToAPIError(tc.inputError)

			require.Equal(t, tc.expectedStatus, result.Status)
			require.Equal(t, tc.expectedMsg, result.Message)
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		status         int
		err            error
		expectedStatus int
		expectedBody   APIError
	}{
		{
			name:           "Standard error",
			status:         http.StatusBadRequest,
			err:            errors.New("test error"),
			expectedStatus: http.StatusBadRequest,
			expectedBody: APIError{
				Status:  http.StatusBadRequest,
				Message: "test error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()

			WriteError(w, tc.status, tc.err)

			require.Equal(t, tc.expectedStatus, w.Code)

			var result APIError
			err := json.NewDecoder(w.Body).Decode(&result)
			require.NoError(t, err)
			require.Equal(t, tc.expectedBody, result)
		})
	}
}

func TestWriteAPIError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		inputError     APIError
		expectedStatus int
		expectedBody   APIError
	}{
		{
			name: "Standard API error",
			inputError: APIError{
				Status:  http.StatusNotFound,
				Message: "Not found",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: APIError{
				Status:  http.StatusNotFound,
				Message: "Not found",
			},
		},
		{
			name: "Internal server error",
			inputError: APIError{
				Status:  http.StatusInternalServerError,
				Message: "Internal error",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: APIError{
				Status:  http.StatusInternalServerError,
				Message: "Internal error",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()

			WriteAPIError(w, tc.inputError)

			require.Equal(t, tc.expectedStatus, w.Code)
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var result APIError
			err := json.NewDecoder(w.Body).Decode(&result)
			require.NoError(t, err)
			require.Equal(t, tc.expectedBody, result)
		})
	}
}

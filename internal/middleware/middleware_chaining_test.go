package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateMiddlewareChain(t *testing.T) {
	t.Parallel()

	t.Run("Empty chain returns original handler", func(t *testing.T) {
		t.Parallel()

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		chain := CreateMiddlewareChain()
		wrapped := chain(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.True(t, handlerCalled)
	})

	t.Run("Single middleware wraps handler", func(t *testing.T) {
		t.Parallel()

		mwCalled := false
		handlerCalled := false

		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mwCalled = true
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		chain := CreateMiddlewareChain(mw)
		wrapped := chain(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.True(t, mwCalled)
		require.True(t, handlerCalled)
	})

	t.Run("Middleware can modify response", func(t *testing.T) {
		t.Parallel()

		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "value")
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		chain := CreateMiddlewareChain(mw)
		wrapped := chain(handler)

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)

		require.Equal(t, "value", rr.Header().Get("X-Test"))
		require.Equal(t, http.StatusOK, rr.Code)
	})
}

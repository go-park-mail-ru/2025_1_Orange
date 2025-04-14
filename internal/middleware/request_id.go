package middleware

import (
	"ResuMatch/internal/utils"
<<<<<<< HEAD
	"github.com/google/uuid"
=======
>>>>>>> a6396a4 (Fix mistakes)
	"net/http"

<<<<<<< HEAD
=======
	"github.com/google/uuid"
)

>>>>>>> a6396a4 (Fix mistakes)
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
			ctx := utils.SetRequestID(r.Context(), requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewString()
				r.Header.Set("X-Request-ID", requestID)
			}

			w.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return requestID
	}
	return ""
}

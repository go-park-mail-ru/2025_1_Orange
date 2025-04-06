package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type ctxKeyRequestID struct{}

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := uuid.NewString()
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

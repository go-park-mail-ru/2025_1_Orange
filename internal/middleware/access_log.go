package middleware

import (
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func AccessLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			cw := &hijackableResponseWriter{ResponseWriter: w}
			next.ServeHTTP(cw, r)

			requestID := utils.GetRequestID(r.Context())

			l.Log.WithFields(logrus.Fields{
				"method":    r.Method,
				"path":      r.URL.Path,
				"status":    cw.statusCode,
				"ip":        getClientIP(r),
				"ua":        r.UserAgent(),
				"latency":   time.Since(start).String(),
				"requestID": requestID,
			}).Info("Request")
		})
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

package middleware

import (
	l "ResuMatch/pkg/logger"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (c *customResponseWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
	c.ResponseWriter.WriteHeader(statusCode)
}

func AccessLogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			cw := &customResponseWriter{ResponseWriter: w}
			next.ServeHTTP(cw, r)

			requestID := r.Header.Get("X-Request-ID")

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

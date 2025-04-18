package middleware

import (
	"ResuMatch/internal/transport/http/utils"
	globalUtils "ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

func RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("%v", err))
					requestID := globalUtils.GetRequestID(r.Context())
					l.Log.WithFields(logrus.Fields{
						"requestID": requestID,
					}).Fatal("Recovery middleware panic")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

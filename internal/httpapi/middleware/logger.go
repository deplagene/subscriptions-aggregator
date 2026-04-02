package middleware

import (
	"net/http"

	"github.com/theartofdevel/logging"
)

func LoggerContext(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxWithLogger := logging.ContextWithLogger(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctxWithLogger))
		})
	}
}

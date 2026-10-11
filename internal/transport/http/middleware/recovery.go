package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"shared-shopping-list-server/internal/transport/http/httputils"
)

func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"panic", rec,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)

					httputils.WriteError(w, http.StatusInternalServerError, "internal_error", "")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

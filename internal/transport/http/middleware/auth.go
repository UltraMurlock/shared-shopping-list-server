package middleware

import (
	"context"
	"net/http"
	"shared-shopping-list-server/internal/auth"
	"shared-shopping-list-server/internal/transport/http/httputils"
	"strings"
	"uuid"
)

type authCtxKey struct{}

var userIDKey authCtxKey

func Auth(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractBearer(r)
			if tokenString == "" {
				httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing token")
				return
			}

			claims, err := jwt.Parse(tokenString)
			if err != nil {
				httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				httputils.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}

	return strings.TrimPrefix(h, prefix)
}

package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type AuthMiddleware struct {
	cookieName string
	tokens     TokenValidator
}

type TokenValidator interface {
	Validate(tokenValue string) (int64, error)
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewAuth(cookieName string, tokenValidator TokenValidator) *AuthMiddleware {
	return &AuthMiddleware{
		cookieName: cookieName,
		tokens:     tokenValidator,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.cookieName)
		if err != nil || cookie.Value == "" {
			writeUnauthorized(w)
			return
		}

		userID, err := m.tokens.Validate(cookie.Value)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	if err := json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"}); err != nil {
		log.Printf("failed to write unauthorized response: %v", err)
	}
}

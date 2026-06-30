package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/panhao/url-shortener/internal/service"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextUsername contextKey = "username"
)

func AuthRequired(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"Missing authorization token"}`, http.StatusUnauthorized)
				return
			}

			userID, username, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextUsername, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthOptional(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token != "" {
				userID, username, err := authService.ValidateToken(token)
				if err == nil {
					ctx := context.WithValue(r.Context(), ContextUserID, userID)
					ctx = context.WithValue(ctx, ContextUsername, username)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

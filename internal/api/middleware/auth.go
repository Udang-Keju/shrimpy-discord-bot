package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the custom claims stored inside the JWT session token.
type Claims struct {
	jwt.RegisteredClaims
	Username      string   `json:"username"`
	Avatar        string   `json:"avatar"`
	ManagedGuilds []string `json:"managed_guilds"`
}

// AuthMiddleware creates a middleware that validates signed JWTs.
// It checks either the Authorization: Bearer token header or a "session" cookie.
func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			// 1. Check Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}

			// 2. Check session cookie
			if tokenString == "" {
				cookie, err := r.Cookie("session")
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				http.Error(w, `{"error": {"code": "UNAUTHORIZED", "message": "Missing authentication session"}}`, http.StatusUnauthorized)
				return
			}

			// Parse and validate JWT
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return jwtSecret, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error": {"code": "UNAUTHORIZED", "message": "Invalid or expired session token"}}`, http.StatusUnauthorized)
				return
			}

			// Inject user info into context using apiutil keys
			ctx := context.WithValue(r.Context(), apiutil.UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, apiutil.ManagedGuildsKey, claims.ManagedGuilds)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

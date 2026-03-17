package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/R4x4nJ031/project-man/internal/config"
	"github.com/R4x4nJ031/project-man/internal/context"
	"github.com/R4x4nJ031/project-man/internal/response"
	jwt "github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {

				if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, errors.New("unexpected signing method")
				}

				return []byte(cfg.JWTSecret), nil
			})

			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(*CustomClaims)
			if !ok || !token.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid claims")
				return
			}

			if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
				response.Error(w, http.StatusUnauthorized, "token expired")
				return
			}

			if claims.Issuer != cfg.JWTIssuer {
				response.Error(w, http.StatusUnauthorized, "invalid issuer")
				return
			}
			validAudience := false
			for _, aud := range claims.Audience {
				if aud == cfg.JWTAud {
					validAudience = true
					break
				}
			}

			if !validAudience {
				response.Error(w, http.StatusUnauthorized, "invalid audience")
				return
			}

			secCtx := &context.SecurityContext{
				UserID: claims.UserID,
				Email:  claims.Email,
			}

			ctx := context.WithSecurityContext(r.Context(), secCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

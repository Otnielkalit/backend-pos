package middleware

import (
	"strings"

	"github.com/Otnielkalit/backend-pos/internal/infrastructure/config"
	"github.com/Otnielkalit/backend-pos/internal/shared/apperror"
	"github.com/Otnielkalit/backend-pos/internal/shared/entity"
	"github.com/Otnielkalit/backend-pos/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth returns a Gin middleware that validates the Bearer JWT token.
// On success, it stores *entity.ActorClaims in the Gin context under entity.ContextKeyActor.
// On failure, it aborts the request with a 401 response.
//
// Usage in route registration:
//
//	r.Use(middleware.Auth(cfg.JWT))
func Auth(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Err(c, apperror.NewUnauthorized("authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Err(c, apperror.NewUnauthorized("authorization header must be Bearer token"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &entity.ActorClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC (prevent algorithm confusion attacks)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperror.NewUnauthorized("unexpected signing method")
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Err(c, apperror.NewUnauthorized("invalid or expired token"))
			c.Abort()
			return
		}

		// Store validated claims in context for downstream handlers
		c.Set(string(entity.ContextKeyActor), claims)
		c.Next()
	}
}

// MustGetActor extracts *entity.ActorClaims from the Gin context.
// Panics if the actor is not set — this should only be called on routes
// protected by the Auth middleware.
func MustGetActor(c *gin.Context) *entity.ActorClaims {
	val, exists := c.Get(string(entity.ContextKeyActor))
	if !exists {
		panic("middleware: MustGetActor called on unprotected route")
	}
	claims, ok := val.(*entity.ActorClaims)
	if !ok {
		panic("middleware: actor claims type assertion failed")
	}
	return claims
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets Cross-Origin Resource Sharing headers.
// Only specific origins are allowed — wildcard "*" is NOT used for security reasons.
//
// Allowed origins are read from the allowedOrigins parameter.
// In development, pass []string{"*"} only if explicitly needed.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowedSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is in allowed list (or wildcard is explicitly passed)
		_, isAllowed := allowedSet[origin]
		_, isWildcard := allowedSet["*"]

		if isWildcard {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

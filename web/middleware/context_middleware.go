package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func ContextMiddleware(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attach the context to the request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

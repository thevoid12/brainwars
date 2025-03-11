package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func CustomProfileMiddleware(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()
	}
}

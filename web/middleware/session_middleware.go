package middleware

import (
	"github.com/gin-gonic/gin"
)

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// we will some how get the user id after authentication and store it in context
		// userID := "00000000-0000-0000-0000-000000000001"
		// ctx = context.WithValue(ctx, constants.CONTEXT_KEY_USER_ID, userID)
		// Attach the context to the request
		// c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ctx := c.Request.Context() // this context has logger in it
		// fmt.Println(c.Request.Method)
		// // if c.Request.Method == http.MethodPost ||
		// // 	c.Request.Method == http.MethodPut ||
		// // 	c.Request.Method == http.MethodDelete {

		// if c.FullPath() == "/bw/home" { // iam seeting the csrf token in /bw/home so there wont be any token
		// 	c.Next()
		// 	return
		// }

		// tokenFromHeader := c.GetHeader("X-XSRF-TOKEN")
		// cookieToken, err := c.Cookie("XSRF-TOKEN")
		// if err != nil || tokenFromHeader == "" || tokenFromHeader != cookieToken {
		// 	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
		// 	return
		// }
		// }
		c.Next()
	}

}

package middleware

import (
	user "brainwars/pkg/users"
	"brainwars/pkg/util"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//	user is validated in the middleware before this
//
// so store the user details from the db into the context so that it can be used accross the app
func CustomProfileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context() // this context has logger in it
		// TODO: somehow pass the userID after authentication
		userID := "00000000-0000-0000-0000-000000000001"
		userInfo, err := user.GetUserDetailsbyID(ctx, uuid.MustParse(userID))
		if err != nil {
			log.Fatalln("get user details by id failed", err)
			return
		}
		ctx = util.SetUserInfoInctx(ctx, userInfo)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

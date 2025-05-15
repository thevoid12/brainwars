package middleware

import (
	user "brainwars/pkg/users"
	"brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"errors"

	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

//	user is validated in the middleware before this
//
// so store the user details from the db into the context so that it can be used accross the app
func CustomProfileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context() // this context has logger in it
		// TODO: somehow pass the userID after authentication
		session := sessions.Default(c)
		profile := session.Get("profile")
		claim := profile.(map[string]interface{})
		sub, err := extractSubFromToken(claim)
		if err != nil {
			log.Fatalln("extract sub from token failed", err)
			return
		}
		username, err := extractNameFromToken(claim)
		if err != nil {
			log.Fatalln("extract name from token failed", err)
			return
		}

		userInfo := util.GetUserInfoFromctx(ctx)
		if userInfo == nil || userInfo.Auth0SubID != sub {
			userInfo, err := user.GetUserDetailsbyAuth0SubID(ctx, sub)
			if err != nil {
				log.Fatalln("get user details by id failed", err)
				return
			}
			if userInfo == nil {
				_, err = user.CreateNewUser(ctx, &model.NewUserReq{
					Auth0SubID: sub,
					UserName:   username,
					UserType:   model.User,
					IsPremium:  false,
				})
				if err != nil {
					return
				}
			}

			ctx = util.SetUserInfoInctx(ctx, userInfo)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// this sub is the unique identyifer for the customer
func extractSubFromToken(claims map[string]interface{}) (string, error) {
	rawSub, ok := claims["sub"]
	if !ok {
		return "", errors.New("sub claim not found")
	}
	sub, ok := rawSub.(string)
	if !ok {
		return "", errors.New("sub claim is not a string")
	}
	return sub, nil
}

func extractNameFromToken(claims map[string]interface{}) (string, error) {
	rawSub, ok := claims["name"]
	if !ok {
		return "", errors.New("name claim not found")
	}
	name, ok := rawSub.(string)
	if !ok {
		return "", errors.New("name claim is not a string")
	}
	return name, nil
}

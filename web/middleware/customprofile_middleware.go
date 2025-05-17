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

		uinfo := session.Get("user_info")

		var userInfo *model.UserInfo

		if uinfo == nil { // session doesnt have the userinfo. we go to the database fecth the info,store it in the session as well as context and use it everywhere
			userInfo, err = user.GetUserDetailsbyAuth0SubID(ctx, sub)
			if err != nil {
				log.Fatalln("get user details by id failed", err)
				return
			}
			if userInfo == nil {
				userInfo, err = user.CreateNewUser(ctx, &model.NewUserReq{
					Auth0SubID: sub,
					UserName:   username,
					UserType:   model.User,
					IsPremium:  false,
				})
				if err != nil {
					return
				}
			}

			session.Set("user_info", userInfo) // gob register in main,go because to set custom go types we need to register the gob beforehand
			err = session.Save()
			if err != nil {
				log.Fatalln("saving userinfo in the session failed", err)
				return
			}

		} else {
			userInfo = uinfo.(*model.UserInfo)
		}

		// i am storing it every single time in context because scope of the data in the context is the http request.
		//Each HTTP request in Go is stateless and independent context.Context lives only for the lifetime of that request.
		// When I navigate to another page ( make a new request), the context starts fresh and does not persist any values from previous requests.
		ctx = util.SetUserInfoInctx(ctx, userInfo)

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

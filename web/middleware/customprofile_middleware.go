package middleware

import (
	rl "brainwars/pkg/rate_limit"
	rlmodel "brainwars/pkg/rate_limit/model"
	user "brainwars/pkg/users"
	"brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"errors"
	"net/http"

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
			log.Print("extract sub from token failed", err)
			session.Clear()
			err = session.Save()
			if err != nil {
				log.Fatalln("saving session after clearing it failed", err)
				return
			}
			//expire the cookie
			c.SetCookie("auth-session", "", -1, "/", "", false, true) // expire the cookie
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "extract sub from token failed"})
			return
		}
		username, err := extractNameFromToken(claim)
		if err != nil {
			session.Clear()
			session.Save()

			//expire the cookie
			c.SetCookie("auth-session", "", -1, "/", "", false, true) // expire the cookie
			log.Print("extract name from token failed", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "extract name from token failed"})
			return
		}

		uinfo := session.Get("user_info")

		var userInfo *model.UserInfo
		if uinfo == nil { // session doesnt have the userinfo. we go to the database fecth the info,store it in the session as well as context and use it everywhere
			userInfo, err = user.GetUserDetailsbyAuth0SubID(ctx, sub)
			if err != nil {
				session.Clear()
				session.Save()

				//expire the cookie
				c.SetCookie("auth-session", "", -1, "/", "", false, true) // expire the cookie
				log.Println("get user details by id failed", err)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "get user details by id failed"})

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

					session.Clear()
					err = session.Save()
					if err != nil {
						log.Fatalln("saving session after clearing it failed", err)
						return
					}
					//expire the cookie
					c.SetCookie("auth-session", "", -1, "/", "", false, true) // expire the cookie
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "create new user failed"})
					return
				}
				err = rl.CreateRateLimit(ctx, rlmodel.RlReq{
					Tries:     0,
					IsPremium: false,
					UserID:    userInfo.ID,
					UserName:  username,
				})
				if err != nil {
					session.Clear()
					session.Save()
					c.SetCookie("auth-session", "", -1, "/", "", false, true) // expire the cookie

					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "setup of rate limiter failed:" + err.Error()})
					return
				}
			}

			session.Set("user_info", userInfo) // gob register in main,go because to set custom go types we need to register the gob beforehand
			err = session.Save()
			if err != nil {
				log.Print("saving userinfo in the session failed", err)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "saving userinfo in the session failed:" + err.Error()})

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

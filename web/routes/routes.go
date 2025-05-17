package routes

import (
	"brainwars/pkg/auth"
	"brainwars/pkg/websocket"
	"brainwars/web/middleware"
	"brainwars/web/ui/handlers"
	assests "brainwars/web/ui/utility"
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(ctx context.Context, l *zap.Logger, auth *auth.Authenticator) (router *gin.Engine) {
	l.Sugar().Info("Initializing logger")

	router = gin.Default()
	router.Use(gin.Recovery())
	// gob is the package that i am using for binary serialization.
	// useful when you need to transmit structured data (like structs) between programs or over the network specific to golang
	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	//Assests and Tailwind
	router.StaticFS("/assets", http.FS(assests.AssestFS)) // Serve embedded files (e.g. JS, CSS, images) under the /assets URL prefix using the embedded filesystem assests.AssestFS.

	router.LoadHTMLGlob("web/ui/templates/*")

	manager := websocket.NewManager(ctx)
	//secure group
	rSecure := router.Group("/bw")
	// middleware
	rSecure.Use(gin.Recovery())
	rSecure.Use(middleware.ContextMiddleware(ctx))
	rSecure.Use(middleware.AuthMiddleware)
	rSecure.Use(middleware.CustomProfileMiddleware())
	rSecure.Use(middleware.SessionMiddleware())

	// login and auth
	//auth group sets the context and calls auth middleware
	rAuth := router.Group("/auth")
	rAuth.Use(middleware.ContextMiddleware(ctx), middleware.AuthMiddleware)

	router.GET("/", handlers.LoginPageHandler)
	router.GET("/login", handlers.LoginHandler(auth))
	router.GET("/callback", handlers.LoginCallbackHandler(auth))
	rSecure.GET("/logout", handlers.LogoutHandler)

	// index
	rSecure.GET("/brainwars", handlers.LandingPageHandler)
	rSecure.GET("/home", handlers.HomeHandler)

	// navbar
	rSecure.GET("/navbar", handlers.GetNavbar)
	// room
	rSecure.POST("/croom", handlers.CreateRoomHandler)
	rSecure.GET("/lroom", handlers.ListAllRoomsHanlder)
	rSecure.GET("/jroom", handlers.JoinRoomHandler)
	rSecure.GET("/ingame/:code", handlers.InGameHandler)
	// http: //localhost:8080/ingame/?roomCode=c5bb492a-051a-42a6-89ec-24e899ea3c14
	// websocket
	rSecure.GET("/ws", manager.ServeWS)

	//questions
	rSecure.GET("/gquest", handlers.GetQuestionHandler)
	rSecure.GET("/quest", handlers.CreateQuestionPageHanlder)
	rSecure.POST("/cquest", handlers.CreateQuestionsHandler)

	// analytics
	rSecure.GET("/analyze/:code", handlers.AnalyticsHandler)
	rSecure.GET("/my-quiz", handlers.MyQuizHistoryHandler)

	for _, route := range router.Routes() {
		l.Sugar().Infof("Route: %s %s", route.Method, route.Path)
	}

	return router
}

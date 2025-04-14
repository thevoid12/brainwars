package routes

import (
	"brainwars/pkg/websocket"
	"brainwars/web/middleware"
	"brainwars/web/ui/handlers"
	assests "brainwars/web/ui/utility"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(ctx context.Context, l *zap.Logger) (router *gin.Engine) {
	l.Sugar().Info("Initializing logger")

	router = gin.Default()
	router.Use(gin.Recovery())
	//Assests and Tailwind
	router.StaticFS("/assets", http.FS(assests.AssestFS)) // Serve embedded files (e.g. JS, CSS, images) under the /assets URL prefix using the embedded filesystem assests.AssestFS.

	router.LoadHTMLGlob("web/ui/templates/*")

	manager := websocket.NewManager(ctx)
	//secure group
	rSecure := router.Group("/bw")

	// middleware
	rSecure.Use(gin.Recovery())
	rSecure.Use(middleware.ContextMiddleware(ctx))
	//rSecure.Use(middleware.AuthMiddleware(ctx))
	rSecure.Use(middleware.CustomProfileMiddleware())
	rSecure.Use(middleware.SessionMiddleware())

	// index
	rSecure.GET("/brainwars", handlers.LandingPageHandler)
	rSecure.GET("/home", handlers.HomeHandler)

	// navbar
	rSecure.GET("/navbar", handlers.GetNavbar)
	// room
	rSecure.POST("/croom", handlers.CreateRoomHandler)
	rSecure.GET("/lroom", handlers.ListAllRoomsHanlder)
	rSecure.GET("/jroom", handlers.JoinRoomHandler)
	rSecure.GET("/quiz", handlers.GameHandler)

	// websocket
	rSecure.GET("/ws", manager.ServeWS)

	//questions
	rSecure.GET("/gquest", handlers.GetQuestionHandler)
	rSecure.GET("/quest", handlers.CreateQuestionPageHanlder)
	rSecure.POST("/cquest", handlers.CreateQuestionsHandler)

	// quiz
	// router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "sec/home") })
	// rSecure.POST("/checkmail", handlers.IndexHandler)
	// router.GET("/test", handlers.IndexHandler) // without middleware
	// router.GET("/", handlers.IndexHandler)
	// router.GET("/about", handlers.AboutHandler)
	//	router.GET("/message", handlers.MessageHandler)

	//auth group sets the context and calls auth middleware
	rAuth := router.Group("/auth")
	rAuth.Use(middleware.ContextMiddleware(ctx), middleware.AuthMiddleware(ctx))
	// rAuth.POST("/gobp/deactivate/:id/:isactive", handlers.IndexHandler)

	for _, route := range router.Routes() {
		l.Sugar().Infof("Route: %s %s", route.Method, route.Path)
	}

	return router
}

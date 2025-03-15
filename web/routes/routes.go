package routes

import (
	"brainwars/pkg/websocket"
	"brainwars/web/middleware"
	assests "brainwars/web/ui/assets"
	"brainwars/web/ui/handlers"
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
	router.StaticFS("/assets", http.FS(assests.AssestFS))

	router.LoadHTMLGlob("web/ui/templates/*")

	manager := websocket.NewManager(ctx)
	//secure group
	rSecure := router.Group("/bw")

	// middleware
	rSecure.Use(gin.Recovery())
	rSecure.Use(middleware.ContextMiddleware(ctx))
	//rSecure.Use(middleware.AuthMiddleware(ctx))
	rSecure.Use(middleware.CustomProfileMiddleware(ctx))
	rSecure.Use(middleware.SessionMiddleware(ctx))

	// index

	rSecure.GET("/brainwars", handlers.LandingPageHandler)
	rSecure.GET("/home", handlers.HomeHandler)
	// room
	rSecure.GET("/room", handlers.CreateRoomPageHandler)
	rSecure.GET("/croom", handlers.CreateRoomHandler) // TODO: This is suppose to be a post req
	rSecure.GET("/lroom", handlers.ListAllRoomsHanlder)
	rSecure.GET("/jroom", handlers.JoinRoomHandler)
	rSecure.GET("/game", handlers.GameHandler)
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

package handlers

import (
	"github.com/gin-gonic/gin"
)

// func IndexHandler(c *gin.Context) {
// 	ctx := c.Request.Context()
// 	l := logs.GetLoggerctx(ctx)
// 	l.Info("this is a test info")
// 	tmpl, err := template.ParseFiles(filepath.Join(viper.GetString("app.uiTemplates"), "index.html"))
// 	if err != nil {
// 		l.Sugar().Errorf("parse template failed", err)
// 		RenderErrorTemplate(c, "Failed to parse form", err)
// 		return
// 	}

// 	// Execute the template and write the output to the response
// 	err = tmpl.Execute(c.Writer, nil)
// 	if err != nil {
// 		l.Sugar().Errorf("execute template failed", err)
// 		return
// 	}
// }

// // IndexHandler handles the home page
// func IndexHandler(c *gin.Context) {
// 	c.HTML(http.StatusOK, "layout.html", gin.H{
// 		"title": "Home Page",
// 	})
// }

// // AboutHandler handles the about page
// func AboutHandler(c *gin.Context) {
// 	files, err := filepath.Glob("web/ui/templates/*")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("**************************")
// 	log.Println("Loaded templates:", files)
// 	c.HTML(http.StatusOK, "about.html", gin.H{
// 		"title": "About Page",
// 	})
// }

// // MessageHandler handles HTMX request for message
// func MessageHandler(c *gin.Context) {
// 	c.String(http.StatusOK, "Hello from the server!")
// }

func LandingPageHandler(c *gin.Context) {
	RenderTemplate(c, "brainwars.html", gin.H{
		"title": "brain war",
	})
}

func HomeHandler(c *gin.Context) {
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

// Room
func CreateRoomPageHandler(c *gin.Context) {

	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func CreateRoomHandler(c *gin.Context) {
	userID := ""
	roomDetails,err := room.CreateRoom(c, userID)
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func ListAllRoomsHanlder(c *gin.Context) {
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func CreateQuestionPageHanlder(c *gin.Context) {
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func CreateQuestionsHandler(c *gin.Context) {
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

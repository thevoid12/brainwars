package handlers

import (
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	"brainwars/pkg/room/model"
	roommodel "brainwars/pkg/room/model"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	fmt.Println("hiiii")
	RenderTemplate(c, "brainwars.html", gin.H{
		"title": "brain war",
	})
}

func HomeHandler(c *gin.Context) {
	fmt.Println("hlo")
	// get the user credentials
	RenderTemplate(c, "home.html", gin.H{
		"title":   "home Page",
		"user-id": "00000000-0000-0000-0000-000000000001",
	})
}

// Room
func CreateRoomPageHandler(c *gin.Context) {

	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func CreateRoomHandler(c *gin.Context) {
	ctx := c.Request.Context() // this context has logger in it

	c.Request.ParseForm()
	fmt.Println(c.Request.Form)
	roomreq := roommodel.RoomReq{
		UserID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Username: "admin",
		UserMeta: "[{}]",
		RoomName: "test room",
		GameType: model.SP,
	}
	joinRoomIDs := []roommodel.UserIDReq{
		{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000002")},
		{UserID: uuid.MustParse("00000000-0000-0000-0000-000000000003")},
	}
	questReq := &quizmodel.QuizReq{
		Topic: "test topic",
		Count: 10,
	}
	err := room.SetupGame(ctx, roomreq, joinRoomIDs, questReq)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to Setup game", err)
	}
	RenderTemplate(c, "home.html", gin.H{
		"title":   "About Page",
		"user-id": "00000000-0000-0000-0000-000000000001",
	})
}

// after the room is created, the user can join the room
// websocket connection is created after the person joins the room
func JoinRoomHandler(c *gin.Context) {
	ctx := c.Request.Context() // this context has logger in it
	// check if there is a room that exists
	roomDetail, err := room.GetRoomByID(ctx, uuid.UUID{})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "unable to join room", err)
	}
	if roomDetail == nil {
		RenderErrorTemplate(c, "home.html", "there is no room", err)
	}

	// check if he has already joined the room if he has then redirect him to the room
	roomMember, err := room.GetRoomMemberByRoomAndUserID(ctx, roommodel.RoomMemberReq{
		UserID: uuid.UUID{},
		RoomID: uuid.UUID{},
	})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to join room", err)
	}
	if roomMember == nil {
		_, err := room.JoinRoomWithRoomCode(ctx, roommodel.RoomMemberReq{
			UserID:   uuid.UUID{},
			RoomID:   uuid.UUID{},
			RoomCode: "",
		})
		if err != nil {
			RenderErrorTemplate(c, "home.html", "Failed to join room", err)
		}
	}

	RenderTemplate(c, "gameroom.html", gin.H{
		"title":  "About Page",
		"roomID": "00000000-0000-0000-0000-000000000005",
		"userID": "00000000-0000-0000-0000-000000000001",
	})
}

func GameHandler(c *gin.Context) {
	RenderTemplate(c, "game.html", gin.H{})
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

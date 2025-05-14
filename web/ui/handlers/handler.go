package handlers

import (
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	"brainwars/pkg/room/model"
	roommodel "brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"fmt"
	"strconv"

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

func GetNavbar(c *gin.Context) {
	RenderSubTemplate(c, "navbar.html", nil)
}

func LandingPageHandler(c *gin.Context) {
	fmt.Println("hiiii")
	RenderTemplate(c, "brainwars.html", gin.H{
		"title": "brain war",
	})
}

func HomeHandler(c *gin.Context) {
	// get the user credentials
	RenderTemplate(c, "home.html", gin.H{
		"title": "home Page",
	})
}

// Room

func CreateRoomHandler(c *gin.Context) {
	ctx := c.Request.Context() // this context has logger in it

	c.Request.ParseForm()

	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID
	gameType := c.PostForm("game-type")
	bots := c.PostFormArray("bots")
	topic := c.PostForm("topic")
	timelimit := c.PostForm("timelimit")
	roomName := c.PostForm("roomName")

	tl, err := strconv.Atoi(timelimit)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "time limit is in wrong format", err)
		return
	}
	questionCount := c.PostForm("questionCount")
	qc, err := strconv.Atoi(questionCount)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "question count is in wrong format", err)
		return
	}

	gt := model.SP
	if gameType == "2" {
		gt = model.MP
	}
	roomreq := roommodel.RoomReq{
		UserID:    userID,
		Username:  "admin",
		UserMeta:  "[{}]",
		RoomName:  roomName,
		GameType:  gt,
		TimeLimit: tl,
	}

	botIDs := []roommodel.UserIDReq{}
	for _, botsInput := range bots {
		botIDs = append(botIDs, roommodel.UserIDReq{UserID: usermodel.BotMap[botsInput]})
	}

	questReq := &quizmodel.QuizReq{
		Topic: topic,
		Count: qc,
	}

	roomCode, err := room.SetupGame(ctx, roomreq, botIDs, questReq)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to Setup game", err)
		return
	}
	if roomreq.GameType == model.SP {
		// redirct to the game room
		c.Redirect(302, fmt.Sprintf("/bw/ingame/%s", roomCode))
		// immidiately join the room,start the game

		return
	}
	// if he is a multiplayer mode then redirect to main page through which he can join the game with room code
	c.Redirect(302, "/bw/home/")

}

// after the room is created, the user can join the room
// websocket connection is created after the person joins the room
func JoinRoomHandler(c *gin.Context) {
	ctx := c.Request.Context() // this context has logger in it
	// check if there is a room that exists
	roomCode := c.PostForm("roomCode")
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID
	roomDetail, err := room.GetRoomByRoomCode(ctx, roomCode)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "unable to join room", err)
	}
	if roomDetail == nil {
		RenderErrorTemplate(c, "home.html", "there is no room", err)
	}

	// check if he has already joined the room if he has then redirect him to the room
	roomMember, err := room.GetRoomMemberByRoomCodeAndUserID(ctx, roommodel.RoomMemberReq{
		UserID:   userID,
		RoomCode: roomCode,
	})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to join room", err)
	}
	if roomMember == nil {
		_, err := room.JoinRoomWithRoomCode(ctx, roommodel.RoomMemberReq{
			UserID:   userID,
			RoomCode: roomCode,
		})
		if err != nil {
			RenderErrorTemplate(c, "home.html", "Failed to join room", err)
		}
		err = room.CreateLeaderBoard(ctx, &model.EditLeaderBoardReq{
			UserID:   userID,
			RoomCode: roomCode,
			Score:    0,
		})
		if err != nil {
			RenderErrorTemplate(c, "home.html", "Failed to setup leaderboard", err)
		}
	}

	RenderTemplate(c, "game.html", gin.H{
		"title":    "game room",
		"roomCode": roomCode,
	})
}

func InGameHandler(c *gin.Context) {
	ctx := c.Request.Context()
	roomCode := c.Param("code")
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	// check if the user is already in the room
	roomMember, err := room.GetRoomMemberByRoomCodeAndUserID(ctx, roommodel.RoomMemberReq{
		UserID:   userID,
		RoomCode: roomCode,
	})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to join room", err)
	}
	if roomMember == nil {
		RenderErrorTemplate(c, "home.html", "you are not in the room", err)
	}

	RenderTemplate(c, "game.html", gin.H{
		"title":    "game room",
		"roomCode": roomCode,
		"userID":   userID,
	})

}

func ListAllRoomsHanlder(c *gin.Context) {
	RenderTemplate(c, "home.html", gin.H{
		"title": "About Page",
	})
}

func GetQuestionHandler(c *gin.Context) {
	// ctx := c.Request.Context()

	// get a generated question

	// after getting the question i display the question
	RenderTemplate(c, "game.html", gin.H{})
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

func AnalyticsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	roomCode := c.Param("code")
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	meta, answers, err := room.ListGameAnalytics(ctx, roommodel.RoomCodeReq{
		UserID:   userID,
		RoomCode: roomCode,
	})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to get analytics", err)
		return
	}
	RenderTemplate(c, "analysis.html", gin.H{
		"title":    "Analytics",
		"roomCode": roomCode,
		"meta":     meta,
		"answers":  answers,
	})
}

func MyQuizHistoryHandler(c *gin.Context) {
	ctx := c.Request.Context()
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	roomDetails, err := room.ListRoom(ctx, roommodel.UserIDReq{
		UserID: userID,
	})
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Failed to get analytics", err)
		return
	}
	RenderTemplate(c, "my_quiz.html", gin.H{
		"title":       "My quiz history",
		"roomDetails": roomDetails,
	})
}

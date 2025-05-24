package handlers

import (
	"brainwars/pkg/auth"
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	"brainwars/pkg/room/model"
	roommodel "brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func LoginPageHandler(c *gin.Context) {
	RenderSubTemplate(c, "login.html", nil)
}

// this is a closure
// A closure is a function plus the variables it remembers from its surrounding context. because gin allows us to pass only one argument
// so we are passing the auth object to the closure
// and then we are using it in the closure
func LoginHandler(authenticator *auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context() // this context has logger in it
		state, err := auth.HandleLogin(ctx, c)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, authenticator.AuthCodeURL(state))
	}
}

// Once users have authenticated using Auth0's Universal Login Page, they will return to the app at the
func LoginCallbackHandler(authenticator *auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if c.Query("state") != session.Get("state") {
			c.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := authenticator.Exchange(c.Request.Context(), c.Query("code"))
		if err != nil {
			c.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		idToken, err := authenticator.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(profile)
		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)
		if err := session.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		c.Redirect(http.StatusTemporaryRedirect, "/bw/home/") // todo: we need to take a look into authenticating the same in auth middleware

	}
}

// Handler for our logout.
func LogoutHandler(c *gin.Context) {
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + c.Request.Host)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	c.SetCookie(
		"auth-session", // name
		"",             // value
		-1,             // maxAge (seconds) — -1 deletes the cookie
		"/",            // path
		"",             // domain
		false,          // secure
		true,           // httpOnly
	)
	c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}

func GetNavbar(c *gin.Context) {
	RenderSubTemplate(c, "navbar.html", nil)
}

func LandingPageHandler(c *gin.Context) {
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

	// RenderSuccessTemplate(c, "home.html", "success testing!")
	// return
	c.Request.ParseForm()

	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID
	gameType := c.PostForm("game-type")
	bots := c.PostFormArray("bots")
	topic := c.PostForm("topic")
	topic = strings.TrimSpace(topic)
	if len(topic) > 50 {
		RenderErrorTemplate(c, "home.html", "length of the topic shouldnt be more than 50 characters", nil)
	}
	if topic == "" {
		RenderErrorTemplate(c, "home.html", "topic cannot be empty", nil)

	}
	timelimit := c.PostForm("timelimit")
	roomName := c.PostForm("roomName")
	difficulty := c.PostForm("difficulty")
	fmt.Println(difficulty)
	tl, err := strconv.Atoi(timelimit)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "time limit is a required field", nil)
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
		Username:  userInfo.UserName,
		UserMeta:  "[{}]",
		RoomName:  roomName,
		GameType:  gt,
		TimeLimit: tl,
	}
	validate := validator.New(validator.WithRequiredStructEnabled())

	// returns nil or ValidationErrors ( []FieldError )
	err = validate.Struct(roomreq)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "invalid user input", err)

	}
	botIDs := []roommodel.UserIDReq{}
	for _, botsInput := range bots {
		botIDs = append(botIDs, roommodel.UserIDReq{UserID: usermodel.BotMap[botsInput]})
	}

	questReq := &quizmodel.QuizReq{
		Topic:      topic,
		Count:      qc,
		Difficulty: quizmodel.Difficulty(difficulty),
	}

	err = validate.Struct(questReq)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "invalid user input", err)

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
	// give a green popup and be in the same page
	RenderSuccessTemplate(c, "home.html", "Successfully room created. Go to My Quiz for your room code")
	// c.Redirect(302, "/bw/home/")

}

// after the room is created, the user can join the room
// websocket connection is created after the person joins the room
func JoinRoomHandler(c *gin.Context) {
	ctx := c.Request.Context() // this context has logger in it
	// check if there is a room that exists
	roomCode := c.Param("code")
	roomCode = strings.TrimSpace(roomCode)
	_, err := uuid.Parse(roomCode) // checking if the room code is a valid uuid
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)

	}
	if roomCode != "" && len(roomCode) > 50 {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)
	}
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID
	roomDetail, err := room.GetRoomByRoomCode(ctx, roomCode)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "unable to join room", nil)
	}
	if roomDetail == nil {
		RenderErrorTemplate(c, "home.html", "there is no room", nil)
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

	c.Redirect(http.StatusPermanentRedirect, "/bw/ingame/"+roomCode)
}

func InGameHandler(c *gin.Context) {
	ctx := c.Request.Context()
	roomCode := c.Param("code")
	roomCode = strings.TrimSpace(roomCode)
	_, err := uuid.Parse(roomCode) // checking if the room code is a valid uuid
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)

	}
	if roomCode != "" && len(roomCode) > 50 {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)
	}

	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	roomDetails, err := room.GetRoomByRoomCode(ctx, roomCode)
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Room Does not exists", nil)
	}

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
		"gameType": roomDetails.GameType,
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
	roomCode = strings.TrimSpace(roomCode)
	_, err := uuid.Parse(roomCode) // checking if the room code is a valid uuid
	if err != nil {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)

	}
	if roomCode != "" && len(roomCode) > 50 {
		RenderErrorTemplate(c, "home.html", "Not a valid room code", nil)
	}
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

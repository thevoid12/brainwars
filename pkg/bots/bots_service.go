package bots

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/quiz"
	"brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	ws "brainwars/pkg/websocket"
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func JoinGameAsBots(ctx context.Context, req roommodel.RoomMemberReq, roomCode string) error {
	l := logs.GetLoggerctx(ctx)
	// check if he has already joined the room if he has then redirect him to the room
	roomMember, err := room.GetRoomMemberByRoomAndUserID(ctx, roommodel.RoomMemberReq{
		UserID: req.UserID,
		RoomID: req.RoomID,
	})
	if err != nil {
		l.Sugar().Error("get room member by room and user ID failed", err)
		return err
	}

	if roomMember == nil {
		_, err := room.JoinRoomWithRoomCode(ctx, roommodel.RoomMemberReq{
			UserID:   roomMember.UserID,
			RoomID:   roomMember.RoomID,
			RoomCode: roomCode,
		})
		if err != nil {
			l.Sugar().Error("join room with room code failed", err)
			return err
		}
	}
	// handlers.RenderTemplate(c, "game.html", gin.H{
	// 	"title":    "game room",
	// 	"roomCode": "8bd9c332-ea09-434c-b439-5b3a39d3de5f",
	// 	"userID":   "00000000-0000-0000-0000-000000000001",
	// })

	return nil
}

func NewBotClient(ctx context.Context, manager *ws.Manager, roomCode string, botType string) *Client {
	// Generate a bot user ID
	botUserID := uuid.New()

	// Create the bot client
	botClient := ws.NewClient(nil, manager, roomCode, true, botType, botUserID)

	// Register bot in database
	roomID, _ := uuid.FromString(roomCode)

	// Create a bot user in your database
	roomMember, err := room.CreateRoomMember(ctx, roommodel.RoomMemberReq{
		UserID:           botUserID,
		RoomID:           roomID,
		RoomMemberStatus: roommodel.ReadyQuiz,
		RoomCode:         roomCode,
		IsBot:            true,
		BotType:          botType,
	})

	// Start bot behavior routine
	go botClient.handleBotBehavior()

	return botClient
}

func (c *ws.Client) handleBotBehavior() {
	if !c.isBot {
		return // Only run for bots
	}

	// Subscribe to room events
	roomEvents := make(chan ws.Event)
	c.manager.registerRoomEventListener(c.roomCode, roomEvents)

	defer func() {
		c.manager.unregisterRoomEventListener(c.roomCode, roomEvents)
		close(roomEvents)
	}()

	var answerTimer *time.Timer

	for event := range roomEvents {
		// Process events for the room
		switch event.Type {
		case "new_question":
			// Cancel any existing timer
			if answerTimer != nil {
				answerTimer.Stop()
			}

			// Schedule bot to answer based on its type
			var delay time.Duration

			switch c.botType {
			case "30sec":
				delay = time.Duration(rand.Intn(25)+5) * time.Second // 5-30 seconds
			case "1min":
				delay = time.Duration(rand.Intn(30)+30) * time.Second // 30-60 seconds
			case "2min":
				delay = time.Duration(rand.Intn(60)+60) * time.Second // 60-120 seconds
			default:
				delay = time.Duration(rand.Intn(30)+5) * time.Second // Default 5-35 seconds
			}

			// Set timer for bot to answer
			answerTimer = time.AfterFunc(delay, func() {
				c.submitRandomAnswer(delay)
			})

		case "game_end":
			// Clean up and exit
			if answerTimer != nil {
				answerTimer.Stop()
			}
			return
		}
	}
}

func (c *ws.Client) submitRandomAnswer(timeSpent time.Duration) {
	ctx := context.Background()

	// Get current question
	question, err := room.GetCurrentQuestion(ctx, c.roomCode)
	if err != nil {
		log.Printf("Bot failed to get current question: %v", err)
		return
	}

	// Select random answer
	answerOptions := question.Options
	selectedIndex := rand.Intn(len(answerOptions))
	selectedAnswer := answerOptions[selectedIndex]

	// Create answer submission
	answerSubmission := struct {
		UserID     string `json:"userId"`
		QuestionID string `json:"questionId"`
		AnswerID   string `json:"answerId"`
		TimeSpent  int    `json:"timeSpent"`
	}{
		UserID:     c.userID.String(),
		QuestionID: question.ID,
		AnswerID:   selectedAnswer.ID,
		TimeSpent:  int(timeSpent.Milliseconds()),
	}

	// Submit the answer
	err = quiz.CreateAnswer(ctx, model.AnswerReq{
		RoomID:         uuid.UUID{},
		UserID:         uuid.UUID{},
		QuestionID:     uuid.UUID{},
		QuestionDataID: uuid.UUID{},
		AnswerOption:   0,
		IsCorrect:      false,
		AnswerTime:     time.Time{},
		CreatedBy:      "",
	})
	if err != nil {
		log.Printf("Bot failed to submit answer: %v", err)
	}
}

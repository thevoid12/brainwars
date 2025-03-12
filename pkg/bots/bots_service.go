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
	// create a client for bot, automatically be ready and start waiting for the question
	return nil
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

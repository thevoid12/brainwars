package model

import (
	"time"

	"github.com/google/uuid"
)

type QuizReq struct {
	Topic string
	Count int
}

type Options struct {
	ID     int
	Option string
}

type QuestionData struct {
	ID       uuid.UUID // unique id for each question
	Question string
	Options  []Options
	Answer   int // option id: which option is correct
}

// QuestionReq represents the request to create a question
type QuestionReq struct {
	RoomID       uuid.UUID
	Topic        string
	QuestionData []*QuestionData
	CreatedBy    string
	Count        int
}

// EditQuestionReq represents the request to update a question
type EditQuestionReq struct {
	ID           uuid.UUID
	Topic        string
	QuestionData []*QuestionData
	UpdatedBy    string
}

// Question represents the question model
type Question struct {
	ID           uuid.UUID
	RoomID       uuid.UUID
	Topic        string
	QuestionData []*QuestionData
	CreatedOn    time.Time
	UpdatedOn    time.Time
	CreatedBy    string
	UpdatedBy    string
}

// AnswerReq represents the request to create an answer
type AnswerReq struct {
	RoomID         uuid.UUID
	UserID         uuid.UUID
	QuestionID     uuid.UUID
	QuestionDataID uuid.UUID
	AnswerOption   int32
	IsCorrect      bool
	AnswerTime     time.Time
	CreatedBy      string
}

// EditAnswerReq represents the request to update an answer
type EditAnswerReq struct {
	ID           uuid.UUID
	AnswerOption int32
	IsCorrect    bool
	AnswerTime   time.Time
	UpdatedBy    string
}

// Answer represents the answer model
type Answer struct {
	ID           uuid.UUID
	RoomID       uuid.UUID
	UserID       uuid.UUID
	QuestionID   uuid.UUID
	AnswerOption int32
	IsCorrect    bool
	AnswerTime   time.Time
	CreatedBy    string
	UpdatedBy    string
}

// GameState to track game progress
type GameState struct {
	RoomCode             string        `json:"roomCode"`
	Status               string        `json:"status"` // "waiting", "in_progress", "ended"
	CurrentRound         int           `json:"currentRound"`
	TotalRounds          int           `json:"totalRounds"`
	Questions            []Question    `json:"questions"`
	Participants         []Participant `json:"participants"`
	StartTime            time.Time     `json:"startTime"`
	CurrentQuestionIndex int           `json:"currentQuestionIndex"`
}

type Participant struct {
	UserID   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	IsBot    bool      `json:"isBot"`
	Score    int       `json:"score"`
	IsReady  bool      `json:"isReady"`
}

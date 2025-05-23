package model

import (
	roommodel "brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"time"

	"github.com/google/uuid"
)

type QuizReq struct {
	Topic      string     `validate:"required"`
	Count      int        `validate:"required"`
	Difficulty Difficulty `validate:"required"`
}

type Difficulty string

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

type Options struct {
	ID     int    `json:"id"`
	Option string `json:"option"`
}

type QuestionData struct {
	ID       uuid.UUID `json:"id"` // unique id for each question
	Question string    `json:"question"`
	Options  []Options `json:"options"`
	Answer   int       `json:"answer"` // option id: which option is correct
}

// QuestionReq represents the request to create a question
type QuestionReq struct {
	RoomCode      string          `validate:"required"`
	Topic         string          `validate:"required"`
	QuestionCount int             `validate:"required"`
	QuestionData  []*QuestionData `validate:"required"`
	CreatedBy     string          `validate:"required"`
	TimeLimit     int             `validate:"required"`
}

// EditQuestionReq represents the request to update a question
type EditQuestionReq struct {
	ID            uuid.UUID
	Topic         string
	QuestionCount int
	QuestionData  []*QuestionData
	UpdatedBy     string
	TimeLimit     int
}

// Question represents the question model
type Question struct {
	ID uuid.UUID
	// RoomID        uuid.UUID
	RoomCode      string
	Topic         string
	QuestionCount int // total number of questions for that room
	QuestionData  []*QuestionData
	TimeLimit     int
	CreatedOn     time.Time
	UpdatedOn     time.Time
	CreatedBy     string
	UpdatedBy     string
}

// AnswerReq represents the request to create an answer
type AnswerReq struct {
	// RoomID         uuid.UUID
	RoomCode       string
	UserID         uuid.UUID `json:"playerID"`
	QuestionID     uuid.UUID
	QuestionDataID uuid.UUID `json:"questionDataID"`
	AnswerOption   int32     `json:"answerOption"`
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
	ID uuid.UUID
	// RoomID       uuid.UUID
	RoomCode       string
	UserID         uuid.UUID
	UserDetails    *usermodel.UserInfo
	QuestionNumber int // just for ui purpose
	QuestionID     uuid.UUID
	QuestionDataID uuid.UUID
	QuestionData   *QuestionData
	AnswerOption   int32
	IsCorrect      bool
	AnswerTime     time.Time
	CreatedBy      string
	UpdatedBy      string
}

// GameState to track game progress
type GameState struct {
	RoomCode             string               `json:"roomCode"`
	RoomStatus           roommodel.RoomStatus `json:"status"` // "waiting", "in_progress", "ended"
	CurrentRound         int                  `json:"currentRound"`
	TotalRounds          int                  `json:"totalRounds"`
	Questions            *Question            `json:"questions"`
	Participants         []Participant        `json:"participants"`
	StartTime            time.Time            `json:"startTime"`
	CurrentQuestionIndex int                  `json:"currentQuestionIndex"`
}

type Participant struct {
	UserID              uuid.UUID `json:"userId"`
	Username            string    `json:"username"`
	IsBot               bool      `json:"isBot"`
	Score               int       `json:"score"`
	Position            int
	IsReady             bool      `json:"isReady"`
	LastAnsweredQestion uuid.UUID `json:"answerID"`
	LastChoosenOption   int       `json:"chosenOption"`
	IsExited            bool      `json:"exited"`
}

type QuizError struct {
	Message string `json:"errorMessage"`
}

type EndGamePayload struct {
	Message      string        `json:"message"`
	Participants []Participant `json:"scores"`
	FinishTime   time.Time     `json:"finishTime"`
}

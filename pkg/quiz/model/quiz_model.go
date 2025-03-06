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
	ID     uuid.UUID
	Option string
}

type QuestionData struct {
	ID       uuid.UUID // unique id for each question
	Question string
	Options  []Options
	Answer   uuid.UUID // option id: which option is correct
}

// QuestionReq represents the request to create a question
type QuestionReq struct {
	RoomID       uuid.UUID
	Topic        string
	QuestionData []*QuestionData
	CreatedBy    uuid.UUID
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

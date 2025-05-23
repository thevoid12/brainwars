package model

import (
	usermodel "brainwars/pkg/users/model"
	"time"

	"github.com/google/uuid"
)

// UserType is a custom type for defining user types
type UserType string

const (
	Bot   UserType = "BOT"
	Human UserType = "HUMAN"
)

type GT string

const (
	SP GT = "SINGLE_PLAYER"
	MP GT = "MULTI_PLAYER"
)

type RoomStatus string

const (
	Started RoomStatus = "STARTED"
	Ended   RoomStatus = "ENDED"
	Waiting RoomStatus = "WAITING" // room is created but the game has not started so people can join in ie waiting for players
	Deleted RoomStatus = "DELETED"
)

// RoomReq is a struct that defines the request body for creating a room
type RoomReq struct {
	UserID    uuid.UUID `validate:"required"`
	Username  string    `validate:"required"`
	UserMeta  string    `validate:"required"`
	RoomName  string    `validate:"required"`
	GameType  GT        `validate:"required"`
	TimeLimit int       `validate:"required"` // max time allocated for each question
}

type RoomMemberStatus string

const (
	CreateQuiz   RoomMemberStatus = "CREATE_QUIZ"
	JoinQuiz     RoomMemberStatus = "JOIN_QUIZ"
	ReadyQuiz    RoomMemberStatus = "READY_QUIZ"
	LeaveQuiz    RoomMemberStatus = "LEAVE_QUIZ"
	KickedQuiz   RoomMemberStatus = "KICKED_QUIZ" // KICKED OUT OF THE ROOM
	BotReadyQuiz RoomMemberStatus = "BOT_READY_QUIZ"
)

// Room is a struct that defines the room model
type Room struct {
	ID            uuid.UUID
	RoomName      string
	RoomCode      string
	UserMeta      string
	RoomMeta      string
	RoomChat      string
	GameType      GT
	Roomstatus    RoomStatus
	IsActive      bool
	IsDeleted     bool
	CreatedBy     string
	UpdatedBy     string
	CreatedOn     time.Time
	UpdatedOn     time.Time
	QuestionTopic string // for listing
	TimeLimit     int    // for listing
}

type EditRoomReq struct {
	ID         uuid.UUID
	UserMeta   string
	RoomName   string
	RoomLock   bool
	GameType   GT
	Roomstatus RoomStatus
}

type RoomMetaReq struct {
	RoomCode string
	RoomMeta string
}

type RoomMember struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	UserDetails      usermodel.UserInfo
	RoomCode         string
	RoomID           uuid.UUID // primary key of room table
	IsBot            bool
	JoinedOn         time.Time
	RoomMemberStatus RoomMemberStatus
	IsActive         bool
	IsDeleted        bool
	CreatedBy        string
	UpdatedBy        string
	CreatedOn        time.Time
	UpdatedOn        time.Time
}

type RoomMemberReq struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RoomMemberStatus RoomMemberStatus
	RoomCode         string
	IsBot            bool
	RoomID           uuid.UUID
}

type UserIDReq struct {
	UserID uuid.UUID
}

/******** Leader board ***************/
type Leaderboard struct {
	ID       uuid.UUID // leaderboard id
	RoomCode string
	UserID   uuid.UUID
	Score    float64
}
type EditLeaderBoardReq struct {
	UserID   uuid.UUID
	RoomCode string
	Score    float64
}

type RoomCodeReq struct {
	UserID   uuid.UUID
	RoomCode string
}

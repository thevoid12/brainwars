package model

import (
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
	Active  RoomStatus = "ACTIVE" // room is created but the game has not started so people can join in
	Deleted RoomStatus = "DELETED"
)

// RoomReq is a struct that defines the request body for creating a room
type RoomReq struct {
	UserID   uuid.UUID
	Username string
	UserMeta string
	RoomName string
	GameType GT
}

type RoomMemberStatus string

const (
	JoinQuiz   RoomMemberStatus = "JOIN_QUIZ"
	ReadyQuiz  RoomMemberStatus = "READY_QUIZ"
	LeaveQuiz  RoomMemberStatus = "LEAVE_QUIZ"
	KickedQuiz RoomMemberStatus = "KICKED_QUIZ" // KICKED OUT OF THE ROOM
)

// Room is a struct that defines the room model
type Room struct {
	ID         uuid.UUID
	RoomName   string
	UserType   string
	UserMeta   string
	RoomMeta   string
	RoomChat   string
	Premium    bool
	GameType   GT
	Roomstatus RoomStatus
	IsActive   bool
	IsDeleted  bool
	CreatedBy  string
	CreatedOn  time.Time
	UpdatedOn  time.Time
}

type EditRoomReq struct {
	ID         uuid.UUID
	UserMeta   string
	RoomName   string
	RoomLock   bool
	GameType   GT
	Roomstatus RoomStatus
}

type RoomMember struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RoomID           uuid.UUID
	RoomCode         string
	UserType         UserType
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
	RoomID           uuid.UUID
	RoomMemberStatus RoomMemberStatus
	RoomCode         string
	IsBot            bool
	BotType         botType,
}

type UserIDReq struct {
	UserID uuid.UUID
}

/******** Leader board ***************/
type Leaderboard struct {
	RoomID uuid.UUID
	UserID uuid.UUID
	Score  float64
}
type EditLeaderBoardReq struct {
	UserID uuid.UUID
	RoomID uuid.UUID
	Score  float64
}

type RoomIDReq struct {
	UserID uuid.UUID
	RoomID uuid.UUID
}

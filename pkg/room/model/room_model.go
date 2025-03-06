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

// RoomReq is a struct that defines the request body for creating a room
type RoomReq struct {
	UserID   uuid.UUID
	Username string
	UserMeta string
	RoomName string
	GameType GT
}

// Room is a struct that defines the room model
type Room struct {
	ID           uuid.UUID
	RoomName     string
	RefreshToken string
	UserType     string
	UserMeta     string
	Premium      bool
	GameType     GT
	IsActive     bool
	IsDeleted    bool
	CreatedBy    uuid.UUID
	CreatedOn    time.Time
	UpdatedOn    time.Time
}

type RoomMemberReq struct {
	UserID uuid.UUID
	RoomID uuid.UUID
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

type RoomLBReq struct {
	UserID uuid.UUID
	RoomID uuid.UUID
}

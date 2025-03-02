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

// RoomReq is a struct that defines the request body for creating a room
type RoomReq struct {
	UserID   uuid.UUID
	Username string
	UserMeta string
	RoomName string
}

// Room is a struct that defines the room model
type Room struct {
	ID           uuid.UUID
	RoomName     string
	RefreshToken string
	UserType     string
	UserMeta     string
	Premium      bool
	IsActive     bool
	IsDeleted    bool
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

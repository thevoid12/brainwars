package model

import "github.com/google/uuid"

type RlReq struct {
	UserID    uuid.UUID
	UserName  string
	Tries     int
	IsPremium bool
}

type EditRlReq struct {
	AllowedAttempts int
	Tries           int
	IsPremium       bool
	IsDeleted       bool
}

type Rl struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	AllowedAttempts int
	Tries           int
	IsPremium       bool
	IsDeleted       bool
}

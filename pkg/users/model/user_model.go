package model

import "github.com/google/uuid"

type BotType string

const (
	Sec10 BotType = "10 sec"
	Sec15 BotType = "15 sec"
	Sec20 BotType = "20 sec"
	Sec30 BotType = "30 sec"
	Sec45 BotType = "30 sec"
	Sec1  BotType = "1 min"
	Sec2  BotType = "2 min"
)

type UserType string

const (
	User = "User"
	Bot  = "Bot"
)

type UserInfo struct {
	ID           uuid.UUID
	UserName     string
	RefreshToken string
	UserType     UserType
	BotType
	IsPremium bool
	IsActive  bool
	IsDeleted bool
}

type NewUserReq struct {
	UserName string
	UserType
	IsPremium bool
}

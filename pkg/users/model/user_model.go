package model

import "github.com/google/uuid"

type BotType string

const (
	Sec10 BotType = "10 sec"
	Sec15 BotType = "15 sec"
	Sec20 BotType = "20 sec"
	Sec30 BotType = "30 sec"
	Sec45 BotType = "45 sec"
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

var BotTypeMap = map[BotType]int{
	Sec10: 10,
	Sec15: 15,
	Sec20: 20,
	Sec30: 30,
	Sec45: 45,
	Sec1:  60,
	Sec2:  120,
}

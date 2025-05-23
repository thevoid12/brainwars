package model

import (
	"time"

	"github.com/google/uuid"
)

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

// this is not the room's status its the status of the user in the room
type UserStatus string

const (
	UserReady        UserStatus = "ready"
	UserJoined       UserStatus = "joined"
	UserQuestionWait UserStatus = "question_wait"
)

type UserInfo struct {
	ID         uuid.UUID
	Auth0SubID string
	UserName   string
	UserType   UserType
	BotType
	IsPremium bool
	IsActive  bool
	IsDeleted bool
}

type NewUserReq struct {
	Auth0SubID string
	UserName   string
	UserType
	IsPremium bool
}

var BotTypeMap = map[BotType]time.Duration{
	Sec10: 10 * time.Second,
	Sec15: 15 * time.Second,
	Sec20: 20 * time.Second,
	Sec30: 30 * time.Second,
	Sec45: 45 * time.Second,
	Sec1:  time.Minute,
	Sec2:  2 * time.Minute,
}

var BotMap = map[string]uuid.UUID{
	"10 sec": uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	"15 sec": uuid.MustParse("00000000-0000-0000-0000-000000000003"),
	"20 sec": uuid.MustParse("00000000-0000-0000-0000-000000000004"),
	"30 sec": uuid.MustParse("00000000-0000-0000-0000-000000000005"),
	"45 sec": uuid.MustParse("00000000-0000-0000-0000-000000000006"),
	"1 min":  uuid.MustParse("00000000-0000-0000-0000-000000000007"),
	"2 min":  uuid.MustParse("00000000-0000-0000-0000-000000000008"),
}

var BotIDMap = map[uuid.UUID]bool{
	uuid.MustParse("00000000-0000-0000-0000-000000000002"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000003"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000004"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000005"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000006"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000007"): true,
	uuid.MustParse("00000000-0000-0000-0000-000000000008"): true,
}

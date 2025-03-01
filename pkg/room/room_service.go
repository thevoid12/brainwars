package room

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room/model"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// CreateRoom is a function that creates a room
func CreateRoom(ctx context.Context, req model.RoomReq) (roomDetails *model.Room, err error) {

	l := logs.GetLoggerctx(ctx)
	members := []model.RoomMembers{}
	roomID := uuid.New()
	members = append(members, model.RoomMembers{
		UserID: req.UserID,
		RoomID: roomID,
	})
	jsonMembers, err := json.Marshal(members)
	if err != nil {
		l.Sugar().Error("error in marshalling room members", err)
		return nil, err
	}
	params := dbal.CreateRoomParams{
		RoomCode:    uuid.New().String(),
		RoomOwner:   req.UserID,
		RoomMembers: []byte(jsonMembers),
		RoomChat:    []byte("[{}]"),
		RoomMeta:    []byte("[{}]"),
		RoomLock:    false,
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   req.UserID.String(),
		UpdatedBy:   req.UserID.String(),
		ID:          roomID,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	// Assuming you have a sqlc function to save the room details to the database
	dBal := dbal.New(dbConn.Db)
	room, err := dBal.CreateRoom(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not create room in database", err)
		return nil, err
	}
	roomDetails = &model.Room{
		ID:           room.ID,
		Username:     room.U.String(), // Assuming RoomOwner is the username
		RefreshToken: uuid.New(),            
		UserType:     string(model.Human),    
		UserMeta:     string(room.RoomMeta),
		Premium:      false, // No equivalent field in CreateRoomParams
		IsActive:     room.IsActive,
		IsDeleted:    room.IsDeleted,
		CreatedOn:    room.CreatedAt,
		UpdatedOn:    room.UpdatedAt,
	}
	return room, nil
	return req, nil
}

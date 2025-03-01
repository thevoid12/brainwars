package room

import (
	"brainwars/pkg/db/dbal"
	"brainwars/pkg/room/model"
	"context"

	"modernc.org/libc/uuid/uuid"
)

// CreateRoom is a function that creates a room
func CreateRoom(c *context.Context, req model.RoomReq) (roomDetails model.Room, err error) {

	member:= []model.RoomMembers{}
	members= append(member, model.RoomMembers{
		UserID: req.UserID,
		RoomID: req.RoomID,
	})
	params := dbal.CreateRoomParams{
		RoomCode:    uuid.New().String(),
		RoomOwner:   req.UserID,
		RoomMembers: req.RoomMembers,
		RoomChat:    req.RoomChat,
		Leaderboard: req.Leaderboard,
		RoomMeta:    req.RoomMeta,
		RoomLock:    req.RoomLock,
		IsActive:    req.IsActive,
		IsDeleted:   req.IsDeleted,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.UpdatedBy,
	}

	// Assuming you have a sqlc function to save the room details to the database

	err = model.CreateRoom(c, req)
	if err != nil {
		return model.RoomReq{}, err
	}

	return req, nil
}

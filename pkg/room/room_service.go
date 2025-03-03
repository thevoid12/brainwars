package room

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room/model"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateRoom is a function that creates a room
func CreateRoom(ctx context.Context, req model.RoomReq) (roomDetails *model.Room, err error) {

	l := logs.GetLoggerctx(ctx)
	members := []model.RoomMemberReq{}
	roomID := uuid.New()
	members = append(members, model.RoomMemberReq{
		UserID: req.UserID,
		RoomID: roomID,
	})
	jsonMembers, err := json.Marshal(members)
	if err != nil {
		l.Sugar().Error("error in marshalling room members", err)
		return nil, err
	}
	params := dbal.CreateRoomParams{
		RoomCode: uuid.New().String(),
		RoomOwner: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		RoomMembers: []byte(jsonMembers),
		RoomChat:    []byte("[{}]"),
		RoomMeta:    []byte("[{}]"),
		RoomLock:    false,
		IsActive:    true,
		IsDeleted:   false,
		CreatedBy:   req.UserID.String(),
		UpdatedBy:   req.UserID.String(),
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
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
		ID:           room.ID.Bytes,
		RoomName:     room.RoomName.String,
		RefreshToken: string(uuid.New().String()),
		UserType:     string(model.Human),
		UserMeta:     string(room.RoomMeta),
		Premium:      false,
		IsActive:     room.IsActive,
		IsDeleted:    room.IsDeleted,
		CreatedOn:    room.CreatedOn.Time,
		UpdatedOn:    room.UpdatedOn.Time,
	}
	return roomDetails, nil
}

func ListRoom(ctx context.Context, req model.UserIDReq) (roomDetails []*model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	rooms, err := dBal.ListRoomByUserID(ctx, pgtype.UUID{
		Bytes: req.UserID,
		Valid: true,
	})
	if err != nil {
		l.Sugar().Error("Could not list rooms in database", err)
		return nil, err
	}
	for _, room := range rooms {
		roomDetails = append(roomDetails, &model.Room{
			ID:           room.ID.Bytes,
			RoomName:     room.RoomName.String,
			RefreshToken: string(uuid.New().String()),
			UserType:     string(model.Human),
			UserMeta:     string(room.RoomMeta),
			Premium:      false,
			IsActive:     room.IsActive,
			IsDeleted:    room.IsDeleted,
			CreatedOn:    room.CreatedOn.Time,
			UpdatedOn:    room.UpdatedOn.Time,
		})
	}
	return roomDetails, nil
}

func JoinRoom(ctx context.Context, req model.RoomMemberReq) (roomDetails *model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	room, err := dBal.GetRoomByID(ctx, pgtype.UUID{
		Bytes: req.RoomID,
		Valid: true,
	})
	if err != nil || len(room) != 0 {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}
	existingMembers := []*model.RoomMemberReq{}
	err = json.Unmarshal(room[0].RoomMembers, &existingMembers)
	if err != nil {
		l.Sugar().Error("Could not unmarshal room members", err)
		return nil, err
	}
	alreadyJoined := false
	for _, member := range existingMembers {
		if member.UserID == req.UserID {
			alreadyJoined = true
			break
		}
	}
	if alreadyJoined {
		return nil, errors.New("user already joined the room")
	}
	existingMembers = append(existingMembers, &model.RoomMemberReq{
		UserID: req.UserID,
		RoomID: req.RoomID,
	})
	jsonMembers, err := json.Marshal(existingMembers)
	if err != nil {
		l.Sugar().Error("error in marshalling room members", err)
		return nil, err
	}
	err = dBal.UpdateRoomByID(ctx, dbal.UpdateRoomByIDParams{
		ID:          room[0].ID,
		RoomName:    room[0].RoomName,
		RoomMembers: jsonMembers,
		RoomChat:    room[0].RoomChat,
		RoomMeta:    room[0].RoomMeta,
		RoomLock:    room[0].RoomLock,
		IsActive:    room[0].IsActive,
		UpdatedBy:   req.UserID.String(),
	})
	if err != nil {
		l.Sugar().Error("Could not update room members in database", err)
		return nil, err
	}

	roomDetails = &model.Room{
		ID:           room[0].ID.Bytes,
		RoomName:     room[0].RoomName.String,
		RefreshToken: string(uuid.New().String()),
		UserType:     string(model.Human),
		UserMeta:     string(room[0].RoomMeta),
		Premium:      false,
		IsActive:     room[0].IsActive,
		IsDeleted:    room[0].IsDeleted,
		CreatedOn:    room[0].CreatedOn.Time,
		UpdatedOn:    room[0].UpdatedOn.Time,
	}
	return roomDetails, nil
}

func LeaveRoom(ctx context.Context, req model.RoomMemberReq) (err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	room, err := dBal.GetRoomByID(ctx, pgtype.UUID{
		Bytes: req.RoomID,

		Valid: true,
	})
	if err != nil || len(room) != 0 {

		l.Sugar().Error("Could not get room by ID in database", err)
		return err
	}
	existingMembers := []*model.RoomMemberReq{}
	err = json.Unmarshal(room[0].RoomMembers, &existingMembers)
	if err != nil {

		l.Sugar().Error("Could not unmarshal room members", err)
		return err
	}
	updatedMembers := []*model.RoomMemberReq{}
	for _, member := range existingMembers {
		if member.UserID != req.UserID {
			updatedMembers = append(updatedMembers, member)
		}
	}
	jsonMembers, err := json.Marshal(updatedMembers)
	if err != nil {
		l.Sugar().Error("error in marshalling room members", err)
		return err
	}
	err = dBal.UpdateRoomByID(ctx, dbal.UpdateRoomByIDParams{
		ID:          room[0].ID,
		RoomName:    room[0].RoomName,
		RoomMembers: jsonMembers,
		RoomChat:    room[0].RoomChat,
		RoomMeta:    room[0].RoomMeta,
		RoomLock:    room[0].RoomLock,
		IsActive:    room[0].IsActive,
		UpdatedBy:   req.UserID.String(),
	})
	if err != nil {
		l.Sugar().Error("Could not update room members in database", err)
		return err
	}
	return nil
}



package bots

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	"context"
)

func JoinGameAsBots(ctx context.Context, req roommodel.RoomMemberReq, roomCode string) error {
	l := logs.GetLoggerctx(ctx)
	// check if he has already joined the room if he has then redirect him to the room
	roomMember, err := room.GetRoomMemberByRoomAndUserID(ctx, roommodel.RoomMemberReq{
		UserID: req.UserID,
		RoomID: req.RoomID,
	})
	if err != nil {
		l.Sugar().Error("get room member by room and user ID failed", err)
		return err
	}

	if roomMember == nil {
		_, err := room.JoinRoomWithRoomCode(ctx, roommodel.RoomMemberReq{
			UserID:   roomMember.UserID,
			RoomID:   roomMember.RoomID,
			RoomCode: roomCode,
		})
		if err != nil {
			l.Sugar().Error("join room with room code failed", err)
			return err
		}
	}
	// create a client for bot, automatically be ready and start waiting for the question
	return nil
}

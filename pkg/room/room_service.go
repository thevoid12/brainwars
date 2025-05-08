package room

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/quiz"
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SetupGame is a function that sets up a game from room creation,member addition,question generation
// TODO: make everything in transaction
func SetupGame(ctx context.Context, req model.RoomReq, botIDs []model.UserIDReq, questReq *quizmodel.QuizReq) (string, error) {
	l := logs.GetLoggerctx(ctx)
	// Create a room
	roomDetails, err := CreateRoom(ctx, req)
	if err != nil {
		l.Sugar().Error("Could not create room", err)
		return "", err
	}

	// Add room members
	for _, membersID := range botIDs {
		// get user details
		_, err = JoinRoom(ctx, model.RoomMemberReq{
			UserID:           membersID.UserID,
			RoomID:           roomDetails.ID,
			RoomMemberStatus: model.ReadyQuiz,
			IsBot:            true,
		})
		if err != nil {
			l.Sugar().Error("Could not join room", err)
			return "", err
		}

		// creating default users leaderboard value to 0
		err = CreateLeaderBoard(ctx, &model.EditLeaderBoardReq{
			UserID:   membersID.UserID,
			RoomCode: roomDetails.RoomCode,
			Score:    0,
		})
		if err != nil {
			return "", err
		}
	}

	// create questions on that topic which llm will generate
	questionData, err := quiz.GenerateQuiz(ctx, &quizmodel.QuizReq{
		Topic: questReq.Topic,
		Count: questReq.Count,
	})
	if err != nil {
		l.Sugar().Error("Could not generate quiz", err)
		return "", err
	}
	// Create question request

	questionReq := quizmodel.QuestionReq{
		RoomCode:      roomDetails.RoomCode,
		Topic:         questReq.Topic,
		QuestionData:  questionData,
		CreatedBy:     string(roomDetails.CreatedBy),
		Count:         questReq.Count,
		QuestionCount: questReq.Count,
		TimeLimit:     req.TimeLimit,
	}

	// Create questions
	err = quiz.CreateQuestion(ctx, questionReq)
	if err != nil {
		l.Sugar().Error("Could not create question", err)
		return "", err
	}

	return roomDetails.RoomCode, nil
}

// CreateRoom is a function that creates a room
func CreateRoom(ctx context.Context, req model.RoomReq) (roomDetails *model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	roomCode := uuid.New() // TODO: FIGURE OUT how to make it a string rather than uuid
	var roomStatus model.RoomStatus
	if req.GameType == model.SP {
		roomStatus = model.Started // no one can join other than active state
	} else {
		roomStatus = model.Waiting
	}

	roomID := uuid.New() // primary key of room
	params := dbal.CreateRoomParams{
		RoomCode: roomCode.String(),
		RoomOwner: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		RoomChat:  []byte("[{}]"),
		RoomMeta:  []byte("[{}]"),
		RoomLock:  false,
		IsActive:  true,
		IsDeleted: false,
		CreatedBy: req.UserID.String(),
		UpdatedBy: req.UserID.String(),
		ID: pgtype.UUID{
			Bytes: roomID,
			Valid: true,
		},
		RoomName: pgtype.Text{
			String: req.RoomName,
			Valid:  req.RoomName != "",
		},
		GameType:   string(req.GameType),
		RoomStatus: string(roomStatus),
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}

	// Start Transaction
	tx, err := dbConn.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Sugar().Error("Could not begin transaction", err)
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx) // Rollback if any error occurs
			l.Sugar().Error("Transaction rolled back due to error", err)
		} else {
			tx.Commit(ctx) // Commit only if there is no error
		}
	}()

	// Use the transaction for DB operations
	dBal := dbal.New(tx)

	room, err := dBal.CreateRoom(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not create room in database", err)
		return nil, err
	}
	roomMemberParams := dbal.CreateRoomMemberParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomCode: string(roomCode.String()),
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		IsBot:            false,
		RoomMemberStatus: string(model.JoinQuiz),
		IsActive:         true,
		IsDeleted:        false,
		CreatedBy:        req.UserID.String(),
		UpdatedBy:        req.UserID.String(),
		RoomID:           roomID.String(),
	}
	_, err = dBal.CreateRoomMember(ctx, roomMemberParams)
	if err != nil {
		l.Sugar().Error("Could not create new room member in database", err)
		return nil, err
	}

	lbParams := dbal.CreatLeaderBoardParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		RoomCode: string(roomCode.String()),
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		Score:     0,
		CreatedBy: req.UserID.String(),
		UpdatedBy: req.UserID.String(),
	}
	err = dBal.CreatLeaderBoard(ctx, lbParams)
	if err != nil {
		l.Sugar().Error("Could not create new leaderboard in database", err)
		return nil, err
	}

	roomDetails = &model.Room{
		ID:         room.ID.Bytes,
		RoomName:   room.RoomName.String,
		UserMeta:   string(room.RoomMeta),
		IsActive:   room.IsActive,
		IsDeleted:  room.IsDeleted,
		CreatedBy:  req.UserID.String(),
		CreatedOn:  room.CreatedOn.Time,
		UpdatedOn:  room.UpdatedOn.Time,
		GameType:   model.GT(room.GameType),
		RoomMeta:   string(room.RoomMeta),
		RoomChat:   string(room.RoomChat),
		Roomstatus: model.RoomStatus(room.RoomStatus),
		RoomCode:   room.RoomCode,
	}
	return roomDetails, nil
}

func UpdateRoom(ctx context.Context, req model.EditRoomReq) (err error) {
	l := logs.GetLoggerctx(ctx)

	room, err := GetRoomByID(ctx, req.ID)
	if err != nil {
		return err
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateRoomByID(ctx, dbal.UpdateRoomByIDParams{
		ID: pgtype.UUID{
			Bytes: room.ID,
			Valid: true,
		},
		RoomName: pgtype.Text{
			String: req.RoomName,
			Valid:  true,
		},
		RoomChat:   []byte(room.RoomChat),
		RoomMeta:   []byte(room.RoomMeta),
		RoomLock:   req.RoomLock,
		IsActive:   room.IsActive,
		UpdatedBy:  "TODO", // TODO: TAKE THE CURRENT USER DETAIL FROM THE CONTEXT
		GameType:   string(req.GameType),
		RoomStatus: string(req.Roomstatus),
	})
	if err != nil {
		l.Sugar().Error("Could not update room by ID in database", err)
		return err
	}

	return nil
}

func UpdateRoomMeta(ctx context.Context, req model.RoomMetaReq) (err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()
	dBal := dbal.New(dbConn.Db)
	room, err := dBal.GetRoomByRoomCode(ctx, req.RoomCode)
	if err != nil || len(room) == 0 {
		l.Sugar().Error("Could not get room by room code in database", err)
		return err
	}

	err = dBal.UpdateRoomMetaByRoomCode(ctx, dbal.UpdateRoomMetaByRoomCodeParams{
		RoomMeta:  []byte(req.RoomMeta),
		RoomCode:  req.RoomCode,
		UpdatedBy: room[0].UpdatedBy,
	})
	if err != nil {
		l.Sugar().Error("Could not update room meta by room code in database", err)
		return err
	}
	return nil
}

func GetRoomByID(ctx context.Context, roomID uuid.UUID) (roomDetails *model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbrecord, err := dBal.GetRoomByID(ctx, pgtype.UUID{
		Bytes: roomID,
		Valid: true,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}

	roomDetails = &model.Room{
		ID:         dbrecord[0].ID.Bytes,
		RoomName:   dbrecord[0].RoomName.String,
		RoomMeta:   string(dbrecord[0].RoomMeta),
		RoomChat:   string(dbrecord[0].RoomChat),
		GameType:   model.GT(dbrecord[0].GameType),
		Roomstatus: model.RoomStatus(dbrecord[0].RoomStatus),
		IsActive:   dbrecord[0].IsActive,
		IsDeleted:  dbrecord[0].IsDeleted,
		CreatedBy:  dbrecord[0].CreatedBy,
		CreatedOn:  dbrecord[0].CreatedOn.Time,
		UpdatedOn:  dbrecord[0].UpdatedOn.Time,
	}
	return roomDetails, nil
}

func GetRoomByRoomCode(ctx context.Context, roomCode string) (roomDetails *model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbrecord, err := dBal.GetRoomByRoomCode(ctx, roomCode)
	if err == nil && dbrecord == nil {
		return nil, nil
	}

	if err != nil {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}

	roomDetails = &model.Room{
		ID:         dbrecord[0].ID.Bytes,
		RoomName:   dbrecord[0].RoomName.String,
		RoomMeta:   string(dbrecord[0].RoomMeta),
		RoomChat:   string(dbrecord[0].RoomChat),
		GameType:   model.GT(dbrecord[0].GameType),
		Roomstatus: model.RoomStatus(dbrecord[0].RoomStatus),
		IsActive:   dbrecord[0].IsActive,
		IsDeleted:  dbrecord[0].IsDeleted,
		CreatedBy:  dbrecord[0].CreatedBy,
		CreatedOn:  dbrecord[0].CreatedOn.Time,
		UpdatedOn:  dbrecord[0].UpdatedOn.Time,
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
			ID:         room.ID.Bytes,
			RoomName:   room.RoomName.String,
			UserMeta:   string(room.RoomMeta),
			GameType:   model.GT(room.GameType),
			IsActive:   room.IsActive,
			IsDeleted:  room.IsDeleted,
			CreatedOn:  room.CreatedOn.Time,
			UpdatedOn:  room.UpdatedOn.Time,
			RoomCode:   room.RoomCode,
			RoomMeta:   string(room.RoomMeta),
			RoomChat:   string(room.RoomChat),
			Roomstatus: model.RoomStatus(room.RoomStatus),
			CreatedBy:  room.CreatedBy,
			UpdatedBy:  room.UpdatedBy,
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

	// Start Transaction
	tx, err := dbConn.Db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Sugar().Error("Could not begin transaction", err)
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx) // Rollback if any error occurs
			l.Sugar().Error("Transaction rolled back due to error", err)
		} else {
			tx.Commit(ctx) // Commit only if there is no error
		}
	}()

	dBal := dbal.New(tx)

	room, err := dBal.GetRoomByID(ctx, pgtype.UUID{
		Bytes: req.RoomID,
		Valid: true,
	})
	if err != nil {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}

	existingMember, err := dBal.GetRoomMemberByRoomCodeAndUserID(ctx, dbal.GetRoomMemberByRoomCodeAndUserIDParams{
		RoomCode: req.RoomCode,
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
	})
	if err != nil || len(existingMember) != 0 {
		l.Sugar().Error("Could not get room member by room and user ID in database", err)
		return nil, err
	}

	if len(existingMember) > 0 {
		return nil, errors.New("user already joined the room")
	}
	_, err = dBal.CreateRoomMember(ctx, dbal.CreateRoomMemberParams{
		ID:               pgtype.UUID{Bytes: uuid.New(), Valid: true},
		RoomCode:         room[0].RoomCode,
		UserID:           pgtype.UUID{Bytes: req.UserID, Valid: true},
		IsBot:            req.IsBot,
		RoomMemberStatus: string(req.RoomMemberStatus),
		IsActive:         true,
		IsDeleted:        false,
		CreatedBy:        req.UserID.String(),
		UpdatedBy:        req.UserID.String(),
		RoomID:           room[0].ID.String(),
	})

	if err != nil {
		l.Sugar().Error("Could not update room members in database", err)
		return nil, err
	}

	roomDetails = &model.Room{
		ID:         room[0].ID.Bytes,
		RoomName:   room[0].RoomName.String,
		UserMeta:   string(room[0].RoomMeta),
		GameType:   model.GT(room[0].GameType),
		IsActive:   room[0].IsActive,
		IsDeleted:  room[0].IsDeleted,
		CreatedOn:  room[0].CreatedOn.Time,
		UpdatedOn:  room[0].UpdatedOn.Time,
		RoomCode:   room[0].RoomCode,
		RoomMeta:   string(room[0].RoomMeta),
		RoomChat:   string(room[0].RoomChat),
		Roomstatus: model.RoomStatus(room[0].RoomStatus),
		CreatedBy:  room[0].CreatedBy,
		UpdatedBy:  room[0].UpdatedBy,
	}

	return roomDetails, nil
}

func ListRoomMembersByRoomCode(ctx context.Context, req model.RoomCodeReq) (roomMembers []*model.RoomMember, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbRecord, err := dBal.ListRoomMembersByRoomCode(ctx, req.RoomCode)
	if err != nil {
		l.Sugar().Error("Could not list room members by room ID in database", err)
		return nil, err
	}
	for _, member := range dbRecord {
		roomMembers = append(roomMembers, &model.RoomMember{
			ID:               member.ID.Bytes,
			UserID:           member.UserID.Bytes,
			IsBot:            member.IsBot,
			RoomMemberStatus: model.RoomMemberStatus(member.RoomMemberStatus),
			IsActive:         member.IsActive,
			IsDeleted:        member.IsDeleted,
			CreatedBy:        member.CreatedBy,
			UpdatedBy:        member.UpdatedBy,
			UserDetails: usermodel.UserInfo{
				ID:           member.ID_2.Bytes,
				UserName:     member.Username,
				RefreshToken: member.RefreshToken, // TODO: only if the user asks we should reviel refresh token else we should hide it securely
				UserType:     usermodel.UserType(member.UserType),
				IsPremium:    member.Premium,
				IsActive:     member.IsActive,
				IsDeleted:    member.IsDeleted_2,
				BotType:      usermodel.BotType(member.BotType.String),
			},
			RoomCode:  member.RoomCode,
			JoinedOn:  member.JoinedOn.Time,
			CreatedOn: member.CreatedOn.Time,
			UpdatedOn: member.UpdatedOn.Time,
			RoomID:    uuid.MustParse(member.RoomID),
		})
	}
	return roomMembers, nil
}

func GetRoomMemberByRoomCodeAndUserID(ctx context.Context, req model.RoomMemberReq) (roomMember *model.RoomMember, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbRecord, err := dBal.GetRoomMemberByRoomCodeAndUserID(ctx, dbal.GetRoomMemberByRoomCodeAndUserIDParams{
		RoomCode: req.RoomCode,
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
	})
	if err == nil && dbRecord == nil {
		return nil, nil
	}
	if err != nil {
		l.Sugar().Error("Could not get room member by room and user ID in database", err)
		return nil, err
	}

	roomMember = &model.RoomMember{
		ID:               dbRecord[0].ID.Bytes,
		UserID:           dbRecord[0].UserID.Bytes,
		IsBot:            dbRecord[0].IsBot,
		RoomMemberStatus: model.RoomMemberStatus(dbRecord[0].RoomMemberStatus),
		IsActive:         dbRecord[0].IsActive,
		IsDeleted:        dbRecord[0].IsDeleted,
		CreatedBy:        dbRecord[0].CreatedBy,
		UpdatedBy:        dbRecord[0].UpdatedBy,
		UserDetails: usermodel.UserInfo{
			ID:           dbRecord[0].ID_2.Bytes,
			UserName:     dbRecord[0].Username,
			RefreshToken: dbRecord[0].RefreshToken, // TODO: only if the user asks we should reviel refresh token else we should hide it securely
			UserType:     usermodel.UserType(dbRecord[0].UserType),
			IsPremium:    dbRecord[0].Premium,
			IsActive:     dbRecord[0].IsActive,
			IsDeleted:    dbRecord[0].IsDeleted_2,
			BotType:      usermodel.BotType(dbRecord[0].BotType.String),
		},
		RoomCode:  dbRecord[0].RoomCode,
		JoinedOn:  dbRecord[0].JoinedOn.Time,
		CreatedOn: dbRecord[0].CreatedOn.Time,
		UpdatedOn: dbRecord[0].UpdatedOn.Time,
		RoomID:    uuid.MustParse(dbRecord[0].RoomID),
	}
	return roomMember, nil
}

// JoinRoomWithRoomCode is a function that joins a member into a room with a room code
// TODO: (V2) implement ability to lock the room so that someone cant join
func JoinRoomWithRoomCode(ctx context.Context, req model.RoomMemberReq) (roomDetails *model.Room, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)

	dbRoom, err := dBal.GetRoomByRoomCode(ctx, req.RoomCode)
	if err != nil {
		l.Sugar().Error("Could not get room by room code in database", err)
		return nil, err
	}

	roomDetails, err = JoinRoom(ctx, model.RoomMemberReq{
		UserID: req.UserID,
		RoomID: dbRoom[0].ID.Bytes,
	})
	if err != nil {
		l.Sugar().Error("Could not join room", err)
		return nil, err
	}

	return roomDetails, nil
}

func UpdateRoomMemberStatusByRoomCodeAndUserID(ctx context.Context, roomCodeReq *model.RoomCodeReq, roomStatus model.RoomMemberStatus) (err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)

	roomDetails, err := dBal.GetRoomByRoomCode(ctx, roomCodeReq.RoomCode)
	if err != nil || len(roomDetails) == 0 {
		l.Sugar().Error("Could not get room member by ID in database", err)
		return err
	}

	err = dBal.UpdateRoomMemberByRoomCodeAndUserID(ctx, dbal.UpdateRoomMemberByRoomCodeAndUserIDParams{
		RoomMemberStatus: string(roomStatus),
		IsActive:         true,
		UpdatedBy:        util.GetUserInfoFromctx(ctx).ID.String(),
		RoomCode:         roomCodeReq.RoomCode,
		UserID:           pgtype.UUID{Bytes: roomCodeReq.UserID, Valid: true},
	})
	if err != nil {
		l.Sugar().Error("Update room member by room code and user ID failed", err)
		return err
	}

	return nil
}

func UpdateRoomMemberStatusByID(ctx context.Context, req model.RoomMemberReq) (err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)

	_, err = dBal.GetRoomMemberByID(ctx, pgtype.UUID{Bytes: req.ID, Valid: true})
	if err != nil {
		l.Sugar().Error("Could not get room member by ID in database", err)
		return err
	}

	err = dBal.UpdateRoomMemberByID(ctx, dbal.UpdateRoomMemberByIDParams{
		ID:               pgtype.UUID{Bytes: req.ID, Valid: true},
		RoomMemberStatus: string(req.RoomMemberStatus),
		IsActive:         true,
		UpdatedBy:        req.UserID.String(),
	})
	if err != nil {
		l.Sugar().Error("Could not join room", err)
		return err
	}

	return nil
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

	err = dBal.UpdateRoomMemberByRoomCodeAndUserID(ctx, dbal.UpdateRoomMemberByRoomCodeAndUserIDParams{
		RoomCode:         req.RoomCode,
		RoomMemberStatus: string(req.RoomMemberStatus),
		IsActive:         false,
		UpdatedBy:        req.UserID.String(),
		UserID:           pgtype.UUID{Bytes: req.UserID, Valid: true},
	})
	if err != nil {
		l.Sugar().Error("Could not uipdate room member by room and userID in database", err)
	}

	return nil
}

/********************** LEADER BOARD **************************************/
func CreateLeaderBoard(ctx context.Context, req *model.EditLeaderBoardReq) error {
	l := logs.GetLoggerctx(ctx)

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}

	dBal := dbal.New(dbConn.Db)

	err = dBal.CreatLeaderBoard(ctx, dbal.CreatLeaderBoardParams{
		RoomCode: req.RoomCode,
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		Score:     float64(req.Score),
		UpdatedBy: req.UserID.String(),
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CreatedBy: req.UserID.String(),
		IsDeleted: false,
	})
	if err != nil {
		l.Sugar().Error("Create leader board for user failed", err)
		return err
	}

	return nil
}

func UpdateLeaderBoard(ctx context.Context, req *model.EditLeaderBoardReq) (err error) {
	l := logs.GetLoggerctx(ctx)

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}

	dBal := dbal.New(dbConn.Db)

	err = dBal.UpdateLeaderBoardScoreByUserIDAndRoomCode(ctx, dbal.UpdateLeaderBoardScoreByUserIDAndRoomCodeParams{
		RoomCode: req.RoomCode,
		UserID: pgtype.UUID{
			Bytes: req.UserID,
			Valid: true,
		},
		Score:     float64(req.Score),
		UpdatedBy: req.UserID.String(),
	})
	if err != nil {
		l.Sugar().Error("Update leader board score by user id and room id failed", err)
		return err
	}

	return nil
}

func ListLeaderBoardByRoomCode(ctx context.Context, req model.RoomCodeReq) (leaderBoard []*model.Leaderboard, err error) {
	l := logs.GetLoggerctx(ctx)

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}

	dBal := dbal.New(dbConn.Db)
	dbRecord, err := dBal.ListLeaderBoardByRoomCode(ctx, req.RoomCode)
	if err != nil {
		l.Sugar().Error("List leaderboard by room id failed", err)
		return nil, err
	}
	for _, lb := range dbRecord {
		leaderBoard = append(leaderBoard, &model.Leaderboard{
			ID:     lb.ID.Bytes,
			UserID: lb.UserID.Bytes,
			Score:  lb.Score,
		})
	}
	return leaderBoard, err
}

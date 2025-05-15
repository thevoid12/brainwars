package user

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/users/model"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateNewUser(ctx context.Context, req *model.NewUserReq) (userDetails *model.UserInfo, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err

	}
	defer dbConn.Db.Close()
	dBal := dbal.New(dbConn.Db)
	err = dBal.CreateNewUser(ctx, dbal.CreateNewUserParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Auth0Sub: pgtype.Text{
			String: req.Auth0SubID,
			Valid:  true,
		},
		Username:  req.UserName,
		UserType:  string(req.UserType),
		BotType:   pgtype.Text{},
		UserMeta:  []byte("[{}]"),
		Premium:   false,
		IsActive:  true,
		IsDeleted: false,
		CreatedBy: req.UserName,
		UpdatedBy: req.UserName,
	})
	if err != nil {
		l.Sugar().Error("create new user failed", err)
		return nil, err
	}

	return nil, nil
}

func GetUserDetailsbyID(ctx context.Context, userID uuid.UUID) (userDetails *model.UserInfo, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbrecord, err := dBal.GetUserDetailsByID(ctx, pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
	if dbrecord == nil && err == nil {
		return nil, nil
	}
	if err != nil {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}

	userDetails = &model.UserInfo{
		ID:         dbrecord[0].ID.Bytes,
		Auth0SubID: dbrecord[0].Auth0Sub.String,
		UserName:   dbrecord[0].Username,
		UserType:   model.UserType(dbrecord[0].UserType),
		IsPremium:  dbrecord[0].Premium,
		IsActive:   dbrecord[0].IsActive,
		IsDeleted:  dbrecord[0].IsDeleted,
	}

	return userDetails, nil
}

func GetUserDetailsbyAuth0SubID(ctx context.Context, sub string) (userDetails *model.UserInfo, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err

	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	dbrecord, err := dBal.GetUserDetailsByAuth0SubID(ctx, pgtype.Text{
		String: sub,
		Valid:  true,
	})
	if dbrecord == nil && err == nil {
		return nil, nil
	}
	if err != nil {
		l.Sugar().Error("Could not get room by auth0 sub ID in database", err)
		return nil, err
	}

	userDetails = &model.UserInfo{
		ID:         dbrecord[0].ID.Bytes,
		Auth0SubID: dbrecord[0].Auth0Sub.String,
		UserName:   dbrecord[0].Username,
		UserType:   model.UserType(dbrecord[0].UserType),
		IsPremium:  dbrecord[0].Premium,
		IsActive:   dbrecord[0].IsActive,
		IsDeleted:  dbrecord[0].IsDeleted,
	}

	return userDetails, nil
}

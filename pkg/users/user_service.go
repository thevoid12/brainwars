package user

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/users/model"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		l.Sugar().Error("Could not get room by ID in database", err)
		return nil, err
	}

	userDetails = &model.UserInfo{
		ID:           dbrecord[0].ID.Bytes,
		UserName:     dbrecord[0].Username,
		RefreshToken: dbrecord[0].RefreshToken,
		UserType:     model.UserType(dbrecord[0].UserType),
		IsPremium:    dbrecord[0].Premium,
		IsActive:     dbrecord[0].IsActive,
		IsDeleted:    dbrecord[0].IsDeleted,
	}

	return userDetails, nil
}

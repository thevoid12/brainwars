package rl

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/rate_limit/model"
	"brainwars/pkg/util"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spf13/viper"
)

// we call this when we set up the customer for the first time
func CreateRateLimit(ctx context.Context, req model.RlReq) error {
	l := logs.GetLoggerctx(ctx)

	params := dbal.CreateRateLimitParams{
		UserID:    req.UserID.String(),
		Allowed:   viper.GetInt32("rl.maxGamesAllowed"),
		Tries:     int32(req.Tries),
		Premium:   req.IsPremium,
		IsDeleted: false,
		CreatedOn: pgtype.Timestamp{
			Time:             time.Now(),
			InfinityModifier: 0,
			Valid:            true,
		},
		UpdatedOn: pgtype.Timestamp{
			Time:             time.Now(),
			InfinityModifier: 0,
			Valid:            true},
		ID:        uuid.New().String(),
		CreatedBy: req.UserName,
		UpdatedBy: req.UserName,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}

	dBal := dbal.New(dbConn.Db)
	err = dBal.CreateRateLimit(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not create ratelimit in database", err)
		return err
	}

	return nil
}

func UpdateRateLimit(ctx context.Context, req model.EditRlReq) error {
	l := logs.GetLoggerctx(ctx)

	userDetails := util.GetUserInfoFromctx(ctx)
	params := dbal.UpdateRateLimitByUserIDParams{
		UserID:    userDetails.ID.String(),
		Allowed:   int32(req.AllowedAttempts),
		Tries:     int32(req.Tries),
		Premium:   req.IsPremium,
		IsDeleted: req.IsDeleted,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateRateLimitByUserID(ctx, params)
	if err != nil {
		l.Sugar().Error("update ratelimit in database", err)
		return err
	}

	return nil
}

func UpdateRateLimitTries(ctx context.Context, tries int) error {
	l := logs.GetLoggerctx(ctx)

	userDetails := util.GetUserInfoFromctx(ctx)
	params := dbal.UpdateRateLimitTriesByUserIDParams{
		UserID:    userDetails.ID.String(),
		Tries:     int32(tries),
		UpdatedBy: userDetails.UserName,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateRateLimitTriesByUserID(ctx, params)
	if err != nil {
		l.Sugar().Error("update ratelimit tries in database", err)
		return err
	}

	return nil
}

func GetRateLimitByUserID(ctx context.Context) (rl *model.Rl, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}

	userDetails := util.GetUserInfoFromctx(ctx)
	dBal := dbal.New(dbConn.Db)
	rlDetails, err := dBal.GetRateLimitByUserID(ctx, userDetails.ID.String())
	if err != nil {
		l.Sugar().Error("Could not get rate limit by user id in database", err)
		return nil, err
	}

	if len(rlDetails) == 0 {
		return nil, errors.New("no record found")
	}
	rl = &model.Rl{
		ID:              uuid.MustParse(rlDetails[0].ID),
		UserID:          uuid.MustParse(rlDetails[0].UserID),
		AllowedAttempts: int(rlDetails[0].Allowed),
		Tries:           int(rlDetails[0].Tries),
		IsPremium:       rlDetails[0].Premium,
		IsDeleted:       rlDetails[0].IsDeleted,
	}

	return rl, nil
}

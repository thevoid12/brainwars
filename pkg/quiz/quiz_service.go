package quiz

import (
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/quiz/model"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateQuestion creates a new question in the database
func CreateQuestion(ctx context.Context, req model.QuestionReq) error {
	l := logs.GetLoggerctx(ctx)

	json.Unmarshal()
	params := dbal.CreateQuestionParams{
		RoomID:       pgtype.UUID{Bytes: req.RoomID, Valid: true},
		Topic:        pgtype.Text{String: req.Topic, Valid: true},
		QuestionData: req.QuestionData,
		CreatedBy:    req.CreatedBy,
		UpdatedBy:    req.CreatedBy,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.CreateQuestion(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not create question in database", err)
		return err
	}

	return nil
}

// UpdateQuestionByID updates a question by its ID
func UpdateQuestionByID(ctx context.Context, req model.EditQuestionReq) error {
	l := logs.GetLoggerctx(ctx)
	params := dbal.UpdateQuestionByIDParams{
		ID:           pgtype.UUID{Bytes: req.ID, Valid: true},
		Topic:        pgtype.Text{String: req.Topic, Valid: true},
		QuestionData: req.QuestionData,
		UpdatedBy:    req.UpdatedBy,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateQuestion(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not update question in database", err)
		return err
	}

	return nil
}

// ListQuestionsByRoomID lists questions by room ID
func ListQuestionsByRoomID(ctx context.Context, roomID uuid.UUID) ([]*model.Question, error) {
	l := logs.GetLoggerctx(ctx)
	var questionDetails []*model.Question
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	questions, err := dBal.ListQuestionsByRoomID(ctx, pgtype.UUID{Bytes: roomID, Valid: true})
	if err != nil {
		l.Sugar().Error("Could not list questions in database", err)
		return nil, err
	}

	for _, question := range questions {
		questionDetails = append(questionDetails, &model.Question{
			ID:           question.ID.Bytes,
			RoomID:       question.RoomID.Bytes,
			Topic:        question.Topic.String,
			QuestionData: string(question.QuestionData),
			CreatedOn:    question.CreatedOn.Time,
			UpdatedOn:    question.UpdatedOn.Time,
			CreatedBy:    question.CreatedBy,
			UpdatedBy:    question.UpdatedBy,
		})
	}

	return questionDetails, nil
}

// CreateAnswer creates a new answer in the database
func CreateAnswer(ctx context.Context, req model.AnswerReq) error {
	l := logs.GetLoggerctx(ctx)
	params := dbal.CreateAnswerParams{
		RoomID:         req.RoomID,
		UserID:         req.UserID,
		QuestionID:     req.QuestionID,
		QuestionDataID: req.QuestionDataID,
		AnswerOption:   req.AnswerOption,
		IsCorrect:      req.IsCorrect,
		AnswerTime:     req.AnswerTime,
		CreatedBy:      req.CreatedBy,
		UpdatedBy:      req.CreatedBy,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.CreateAnswer(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not create answer in database", err)
		return err
	}

	return nil
}

// UpdateAnswer updates an existing answer in the database
func UpdateAnswer(ctx context.Context, req model.AnswerUpdateReq) error {
	l := logs.GetLoggerctx(ctx)
	params := dbal.UpdateAnswerParams{
		ID:           pgtype.UUID{Bytes: req.ID, Valid: true},
		AnswerOption: req.AnswerOption,
		IsCorrect:    req.IsCorrect,
		AnswerTime:   req.AnswerTime,
		UpdatedBy:    req.UpdatedBy,
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateAnswer(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not update answer in database", err)
		return err
	}

	return nil
}

// ListAnswersByRoomID lists answers by room ID
func ListAnswersByRoomID(ctx context.Context, roomID uuid.UUID) ([]*model.Answer, error) {
	l := logs.GetLoggerctx(ctx)
	var answerDetails []*model.Answer
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	answers, err := dBal.ListAnswersByRoomID(ctx, pgtype.UUID{Bytes: roomID, Valid: true})
	if err != nil {
		l.Sugar().Error("Could not list answers in database", err)
		return nil, err
	}

	for _, answer := range answers {
		answerDetails = append(answerDetails, &model.Answer{
			ID:           answer.ID.Bytes,
			RoomID:       answer.RoomID.Bytes,
			UserID:       answer.UserID.Bytes,
			QuestionID:   answer.QuestionID.Bytes,
			AnswerOption: answer.AnswerOption,
			IsCorrect:    answer.IsCorrect,
			AnswerTime:   answer.AnswerTime.Time,
			CreatedBy:    answer.CreatedBy,
			UpdatedBy:    answer.UpdatedBy,
		})
	}

	return answerDetails, nil
}

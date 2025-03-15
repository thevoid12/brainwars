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

// llm generates the questions based on the topic and count
// and returns the questions in the form of QuestionData
// which is a slice of QuestionData
func GenerateQuiz(ctx context.Context, req *model.QuizReq) (questData []*model.QuestionData, err error) {
	questData = []*model.QuestionData{}
	// sample
	questData = append(questData, &model.QuestionData{
		ID:       uuid.New(),
		Question: "this is test question 1",
		Options:  []model.Options{{ID: 1, Option: "ans 1"}, {ID: 2, Option: "ans 2"}, {ID: 3, Option: "ans 3"}, {ID: 4, Option: "ans 4"}},
		Answer:   1,
	})
	questData = append(questData, &model.QuestionData{
		ID:       uuid.New(),
		Question: "this is test question 2",
		Options:  []model.Options{{ID: 1, Option: "ans 1"}, {ID: 2, Option: "ans 2"}, {ID: 3, Option: "ans 3"}, {ID: 4, Option: "ans 4"}},
		Answer:   2,
	})
	return questData, nil
}

// CreateQuestion creates a new question in the database
func CreateQuestion(ctx context.Context, req model.QuestionReq) error {
	l := logs.GetLoggerctx(ctx)

	quesJson, err := json.Marshal(req.QuestionData)
	if err != nil {
		l.Sugar().Error("Could not marshal question data", err)
		return err
	}

	params := dbal.CreateQuestionParams{
		Topic:         pgtype.Text{String: req.Topic, Valid: true},
		QuestionData:  quesJson,
		CreatedBy:     req.CreatedBy,
		UpdatedBy:     req.CreatedBy,
		QuestionCount: int32(req.QuestionCount),
		RoomCode:      req.RoomCode.String(),
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		TimeLimit:     int32(req.TimeLimit),
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

	quesJson, err := json.Marshal(req.QuestionData)
	if err != nil {
		l.Sugar().Error("Could not marshal question data", err)
		return err
	}

	params := dbal.UpdateQuestionByIDParams{
		ID:            pgtype.UUID{Bytes: req.ID, Valid: true},
		Topic:         pgtype.Text{String: req.Topic, Valid: true},
		QuestionData:  quesJson,
		UpdatedBy:     req.UpdatedBy,
		QuestionCount: int32(req.QuestionCount),
		TimeLimit:     int32(req.TimeLimit),
	}

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	err = dBal.UpdateQuestionByID(ctx, params)
	if err != nil {
		l.Sugar().Error("Could not update question in database", err)
		return err
	}

	return nil
}

// ListQuestionsByRoomCode lists questions by room Code
func ListQuestionsByRoomCode(ctx context.Context, roomCode string) (questionDetails *model.Question, err error) {
	l := logs.GetLoggerctx(ctx)
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	question, err := dBal.GetQuestionsByRoomCode(ctx, roomCode)
	if err != nil {
		l.Sugar().Error("Could not list questions in database", err)
		return nil, err
	}

	questionDetails = &model.Question{
		ID:            question.ID.Bytes,
		RoomCode:      question.RoomCode,
		Topic:         question.Topic.String,
		QuestionData:  nil,
		CreatedOn:     question.CreatedOn.Time,
		UpdatedOn:     question.UpdatedOn.Time,
		CreatedBy:     question.CreatedBy,
		UpdatedBy:     question.UpdatedBy,
		QuestionCount: int(question.QuestionCount),
	}

	qs := []*model.QuestionData{}
	err = json.Unmarshal(question.QuestionData, &qs)
	if err != nil {
		l.Sugar().Error("Could not unmarshal question data", err)
		return nil, err
	}

	questionDetails.QuestionData = qs

	return questionDetails, nil
}

// CreateAnswer creates a new answer in the database
func CreateAnswer(ctx context.Context, req model.AnswerReq) error {
	l := logs.GetLoggerctx(ctx)
	params := dbal.CreateAnswerParams{
		RoomCode:       req.RoomCode,
		UserID:         pgtype.UUID{Bytes: req.UserID, Valid: true},
		QuestionID:     pgtype.UUID{Bytes: req.QuestionID, Valid: true},
		QuestionDataID: pgtype.UUID{Bytes: req.QuestionDataID, Valid: true},
		AnswerOption:   req.AnswerOption,
		IsCorrect:      req.IsCorrect,
		AnswerTime:     pgtype.Timestamp{Time: req.AnswerTime, Valid: true},
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
func UpdateAnswer(ctx context.Context, req model.EditAnswerReq) error {
	l := logs.GetLoggerctx(ctx)
	params := dbal.UpdateAnswerParams{
		ID:           pgtype.UUID{Bytes: req.ID, Valid: true},
		AnswerOption: req.AnswerOption,
		IsCorrect:    req.IsCorrect,
		AnswerTime:   pgtype.Timestamp{Time: req.AnswerTime, Valid: true},
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

// ListAnswersByRoomCode lists answers by room Code
func ListAnswersByRoomCode(ctx context.Context, roomCode string) ([]*model.Answer, error) {
	l := logs.GetLoggerctx(ctx)
	var answerDetails []*model.Answer
	dbConn, err := dbpkg.InitDB()
	if err != nil {
		l.Sugar().Error("Could not initialize database", err)
		return nil, err
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	answers, err := dBal.ListAnswersByRoomCode(ctx, roomCode)
	if err != nil {
		l.Sugar().Error("Could not list answers in database", err)
		return nil, err
	}

	for _, answer := range answers {
		answerDetails = append(answerDetails, &model.Answer{
			ID:       answer.ID.Bytes,
			RoomCode: answer.RoomCode,
			//	RoomID:       answer.RoomID.Bytes,
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

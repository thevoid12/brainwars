package main

import (
	"context"
	"fmt"
	dbpkg "brainwars/pkg/db"
	"brainwars/pkg/db/dbal"
	logs "brainwars/pkg/logger"
	"brainwars/web/routes"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("there is a error loading environment variables", err)
		return
	}
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("config/") // path to look for the config file in

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("there is a error in the path of config file", err)
		} else {
			// Config file was found but another error was produced
			log.Println("error laoding config file from viper", err)
		}
	}

	l, err := logs.InitializeLogger()
	if err != nil {
		log.Println("error initializing logger", err)
	}

	ctx := context.Background()
	ctx = logs.SetLoggerctx(ctx, l)

	l.Sugar().Info("cache initialized successfully")

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer dbConn.Db.Close()

	dBal := dbal.New(dbConn.Db)
	_, err = dBal.CreateTest(ctx, dbal.CreateTestParams{
		Name: "test-void",
		Bio: pgtype.Text{
			String: "this is a test",
			Valid:  true,
		},
	})
	if err != nil {
		l.Sugar().Info("failed", err)
	}
	result, err := dBal.ListTest(ctx)
	if err != nil {
		l.Sugar().Info("failed")
	}
	fmt.Println(result)
	route := routes.Initialize(ctx, l)
	route.Run(":" + viper.GetString("app.port"))
}

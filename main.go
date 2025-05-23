package main

import (
	"brainwars/pkg/auth"
	dbpkg "brainwars/pkg/db"
	logs "brainwars/pkg/logger"
	"brainwars/pkg/users/model"
	"brainwars/web/routes"
	"context"
	"encoding/gob"
	"log"

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

	gob.Register(&model.UserInfo{}) // Register pointer type

	dbConn, err := dbpkg.InitDB()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}
	defer dbConn.Db.Close()

	authenticator, err := auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	route := routes.Initialize(ctx, l, authenticator)
	route.Run(":" + viper.GetString("app.port"))
}

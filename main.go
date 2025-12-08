package main

import (
	"fmt"
	"github.com/WedhaWS/uasgosmt5/config"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLog(viperConfig)
	app := config.NewFiber(viperConfig)
	postgres := config.PostgresConnect(viperConfig, log)
	mongo := config.MongoConnect(viperConfig, log)
	redis := config.NewRedisClient(viperConfig, log)
	validate := config.NewValidator()

	config.Bootstrap(&config.BootstrapConfig{
		Postgres: postgres,
		App:      app,
		Log:      log,
		Config:   viperConfig,
		MongoDB:  mongo,
		Redis:    redis,
		Validate: validate,
	})

	port := viperConfig.GetInt("app.port")
	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}
}
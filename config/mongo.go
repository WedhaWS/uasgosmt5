package config

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoConnect(config *viper.Viper, logs *logrus.Logger) *mongo.Database {

	MongoPort := config.GetString("database.mongodb.port")
	MongoHost := config.GetString("database.mongodb.host")
	clientOptions := options.Client().ApplyURI("mongodb://" + MongoHost + ":" + MongoPort).SetMaxPoolSize(10).SetMinPoolSize(5).SetMaxConnIdleTime(10 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	logs.Info("Connected to MongoDB")

	return client.Database(viper.GetString("database.mongodb.dbname"))
}
package config

import (
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRedisClient(viper *viper.Viper, log *logrus.Logger) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		// Perhatikan: "hots" harusnya "host" (perlu dikoreksi)
		Addr: viper.GetString("database.redis.host") + ":" + viper.GetString("database.redis.port"),
		DB:   viper.GetInt("database.redis.DB"), // Perbaikan dari "redis.db"
	})

	return redisClient
}
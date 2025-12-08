package config

import (
	"database/sql"
	"github.com/WedhaWS/uasgosmt5/app/repository"
	"github.com/WedhaWS/uasgosmt5/app/service"
	"github.com/WedhaWS/uasgosmt5/middleware"
	"github.com/WedhaWS/uasgosmt5/routes"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

type BootstrapConfig struct {
	App      *fiber.App
	Postgres *sql.DB
	MongoDB  *mongo.Database
	Redis    *redis.Client
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(config *BootstrapConfig) {

	//Setup Repository
	UserRepository := repository.NewUserRepository(config.Postgres, config.Log)
	LogoutRepository := repository.NewLogoutRepository(config.Redis, config.Log)

	secret := []byte(config.Config.GetString("app.jwt-secret"))
	//Setup Service
	UserService := service.NewAuthService(UserRepository, LogoutRepository, config.Log, secret)

	RouteConfig := routes.RouteConfig{
		App:            config.App,
		UserService:    UserService,
		AuthMiddleware: middleware.AuthRequired(secret),
	}

	RouteConfig.Setup()

}
package config

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func PostgresConnect(config *viper.Viper, logs *logrus.Logger) *sql.DB {
	PostgresPort := config.GetString("database.postgre.port")
	PostgresHost := config.GetString("database.postgre.host")
	PostgresUser := config.GetString("database.postgre.user")
	PostgresPassword := config.GetString("database.postgre.password")
	PostgresDBName := config.GetString("database.postgre.dbname")

	connStr := "postgres://" + PostgresUser + ":" + PostgresPassword + "@" + PostgresHost + ":" + PostgresPort + "/" + PostgresDBName + "?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	logs.Info("Connected to PostgresSQL")
	return db
}
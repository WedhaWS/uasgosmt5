package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/viper"
)

func PostgreConnect(config *viper.Viper) *sql.DB {
	PostgrePort := config.GetString("database.postgre.port")
	PostgreHost := config.GetString("database.postgre.host")
	PostgreUser := config.GetString("database.postgre.user")
	PostgrePassword := config.GetString("database.postgre.password")
	PostgreDBName := config.GetString("database.postgre.dbname")

	connStr := "postgres://" + PostgreUser + ":" + PostgrePassword + "@" + PostgreHost + ":" + PostgrePort + "/" + PostgreDBName + "?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(fmt.Errorf("failed to open database connection: %w", err))
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

//migrate -database "postgres://artha:passwordku@localhost:5432/prisma_db?sslmode=disable" -path db/migrations_postgre up
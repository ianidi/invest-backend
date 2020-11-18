package db

import (
	"log"

	"github.com/spf13/viper"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx/types"
)

var DB *sqlx.DB

// Init connection to the database
func Init() *sqlx.DB {
	if viper.GetString("connection") == "" {
		log.Fatalln("Please pass the connection string using the -connection option")
	}

	db, err := sqlx.Connect("pgx", viper.GetString("connection"))
	if err != nil {
		log.Fatalf("Unable to establish connection: %v\n", err)
	}

	DB = db
	return DB
}

// GetDB - Using this function to get a connection, you can create your connection pool here
func GetDB() *sqlx.DB {
	return DB
}

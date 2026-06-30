package db

import (
	"2say/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(conf *config.Config) *sql.DB {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.DBPort, conf.DBUser, conf.DBPassword, conf.DBName, conf.SslMode)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil{
		log.Fatalf("Error loading db: %s", err)
	}
	err = db.Ping()
	if err !=nil{
		log.Fatalf("Error ping db: %s", err)
	}
	log.Println("Successfully connected db")

	return db
}

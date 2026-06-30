package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)
// todo значения по умолчанию
type Config struct{
	Host string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
	SslMode string

}

func LoadConfig() *Config{
	err := godotenv.Load("../../.env")
	if err != nil{
		log.Fatal("Error loading .env file")
	}

	log.Println("Successfully loaded config")

	return & Config{
	Host : os.Getenv("DB_HOST"),
	DBPort : os.Getenv("DB_PORT"),
	DBUser : os.Getenv("DB_USER"),
	DBPassword : os.Getenv("DB_PASSWORD"),
	DBName : os.Getenv("DB_NAME"),
	SslMode : os.Getenv("DB_SSLMODE"),
	}
}
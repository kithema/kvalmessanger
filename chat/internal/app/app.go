package app

import (
	"2say/internal/config"
	"2say/internal/db"
	"2say/internal/routes"
	"log"
	"net/http"
)
func Run(){
	
	config := config.LoadConfig()
	db := db.InitDB(config)
	router := routes.SetupRoutes(db)


	log.Fatal(http.ListenAndServe(":8080", router))

}
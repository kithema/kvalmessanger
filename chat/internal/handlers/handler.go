package handlers

import(
	_"net/http"
	"database/sql"
)

type Handler struct{
	DB *sql.DB
}

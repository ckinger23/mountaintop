package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
	"net/http"
)

func SubmitPickHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement DynamoDB put operation
		utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
	}
}

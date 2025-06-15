package handlers

import (
	"net/http"
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
)

func SubmitPickHandler(dbClient *db.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement DynamoDB put operation
		utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
	}
}

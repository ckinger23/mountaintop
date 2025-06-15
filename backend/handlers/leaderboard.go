package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
	"net/http"
)

func GetLeaderboardHandler(dbClient *db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement DynamoDB query
		utils.RespondWithJSON(w, http.StatusOK, []interface{}{})
	}
}

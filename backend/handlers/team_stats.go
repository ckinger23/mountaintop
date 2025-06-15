package handlers

import (
	"net/http"
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
)

func GetTeamStatsHandler(dbClient *db.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement DynamoDB query
		utils.RespondWithJSON(w, http.StatusOK, []interface{}{})
	}
}

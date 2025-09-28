package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/evevioletrose-hash/idle-dungeon/internal/game"
)

// PlayerHandler handles HTTP requests for player data retrieval.
// It returns player information in JSON format for API consumers.
func PlayerHandler(gameServer *game.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := r.URL.Query().Get("id")
		if playerID == "" {
			http.Error(w, "Player ID required", http.StatusBadRequest)
			return
		}

		player := gameServer.GetOrCreatePlayer(playerID)
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(player); err != nil {
			http.Error(w, "Failed to encode player data", http.StatusInternalServerError)
		}
	}
}

// UpgradeHandler handles HTTP POST requests for factory station upgrades.
// It processes upgrade requests and returns updated player data.
func UpgradeHandler(gameServer *game.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		playerID := r.URL.Query().Get("playerID")
		station := r.URL.Query().Get("station")
		
		if playerID == "" || station == "" {
			http.Error(w, "PlayerID and station required", http.StatusBadRequest)
			return
		}

		player := gameServer.GetOrCreatePlayer(playerID)
		success := gameServer.UpgradeStation(player, station)
		
		if !success {
			http.Error(w, "Upgrade failed - insufficient funds or invalid station", http.StatusBadRequest)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(player); err != nil {
			http.Error(w, "Failed to encode player data", http.StatusInternalServerError)
		}
	}
}
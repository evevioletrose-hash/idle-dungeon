// Package handlers contains HTTP and WebSocket request handlers for the idle dungeon game.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/evevioletrose-hash/idle-dungeon/internal/game"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections for real-time multiplayer functionality.
// It upgrades HTTP connections to WebSocket and manages client communication.
func WebSocketHandler(gameServer *game.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := gameServer.GetUpgrader()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Get or generate player ID
		playerID := r.URL.Query().Get("playerID")
		if playerID == "" {
			playerID = generatePlayerID()
		}

		// Get or create player and register connection
		player := gameServer.GetOrCreatePlayer(playerID)
		gameServer.AddClient(conn, player)
		defer gameServer.RemoveClient(conn)

		// Send initial game state to the newly connected client
		initialState, _ := json.Marshal(map[string]interface{}{
			"type":   "gameState",
			"player": player,
		})
		gameServer.BroadcastToClient(conn, initialState)

		// Handle incoming messages from the client
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err == nil {
				handleClientMessage(gameServer, conn, msg)
			}
		}
	}
}

// handleClientMessage processes messages received from WebSocket clients.
// It handles different message types like upgrade requests.
func handleClientMessage(gameServer *game.Server, conn *websocket.Conn, msg map[string]interface{}) {
	player := gameServer.GetPlayerByConnection(conn)
	if player == nil {
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "upgrade":
		station, ok := msg["station"].(string)
		if ok {
			gameServer.UpgradeStation(player, station)
		}
	}
}

// generatePlayerID creates a unique identifier for new players.
// It uses the current timestamp in base36 encoding for uniqueness.
func generatePlayerID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
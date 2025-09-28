// Package main is the entry point for the Idle Dungeon game server.
// It sets up the HTTP server, WebSocket endpoints, and starts the game services.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/evevioletrose-hash/idle-dungeon/internal/game"
	"github.com/evevioletrose-hash/idle-dungeon/internal/handlers"
)

func main() {
	// Initialize the game server
	gameServer := game.NewServer()
	
	// Start the game server background processes
	gameServer.Start()

	// Setup HTTP routes
	setupRoutes(gameServer)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🏰 Idle Dungeon server starting on port %s", port)
	log.Printf("🌐 Game available at http://localhost:%s", port)
	
	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// setupRoutes configures all HTTP endpoints for the game server.
func setupRoutes(gameServer *game.Server) {
	// Serve static files (HTML, CSS, JavaScript)
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	
	// WebSocket endpoint for real-time multiplayer communication
	http.HandleFunc("/ws", handlers.WebSocketHandler(gameServer))
	
	// REST API endpoints
	http.HandleFunc("/api/player", handlers.PlayerHandler(gameServer))
	http.HandleFunc("/api/upgrade", handlers.UpgradeHandler(gameServer))
	
	log.Println("📡 Routes configured:")
	log.Println("  GET  /           - Game web interface")
	log.Println("  WS   /ws         - WebSocket for real-time updates") 
	log.Println("  GET  /api/player - Player data API")
	log.Println("  POST /api/upgrade- Factory upgrade API")
}
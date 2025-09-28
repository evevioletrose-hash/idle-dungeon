// Package game contains the core game logic and server functionality.
package game

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/evevioletrose-hash/idle-dungeon/internal/models"
	"github.com/gorilla/websocket"
)

// Server manages the game state and handles multiplayer connections.
// It processes the game loop, manages WebSocket connections, and broadcasts updates.
type Server struct {
	gameState *models.GameState                   // Central game state containing all players
	clients   map[*websocket.Conn]*models.Player // Map of WebSocket connections to players
	broadcast chan []byte                         // Channel for broadcasting messages to all clients
	register  chan *websocket.Conn               // Channel for registering new client connections
	upgrader  websocket.Upgrader                 // WebSocket upgrader for HTTP connections
	mutex     sync.RWMutex                       // Mutex for thread-safe access to clients map
}

// NewServer creates and initializes a new game server.
func NewServer() *Server {
	return &Server{
		gameState: models.NewGameState(),
		clients:   make(map[*websocket.Conn]*models.Player),
		broadcast: make(chan []byte),
		register:  make(chan *websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// Start begins the game server operations including the game loop and message handling.
func (s *Server) Start() {
	go s.gameLoop()
	go s.handleMessages()
}

// gameLoop runs continuously to process all players and broadcast updates.
// It ticks every second to simulate the idle game progression.
func (s *Server) gameLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		players := s.gameState.GetAllPlayers()
		
		// Process each player's battle
		for _, player := range players {
			s.processPlayer(player)
		}

		// Broadcast updates to all connected clients
		gameUpdate, _ := json.Marshal(map[string]interface{}{
			"type":    "update",
			"players": players,
		})
		
		select {
		case s.broadcast <- gameUpdate:
		default:
		}
	}
}

// handleMessages manages the broadcasting of messages to all connected clients.
func (s *Server) handleMessages() {
	for {
		select {
		case message := <-s.broadcast:
			s.mutex.RLock()
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(s.clients, client)
				}
			}
			s.mutex.RUnlock()
		}
	}
}

// GetOrCreatePlayer retrieves an existing player or creates a new one if not found.
// This method is thread-safe and handles player initialization.
func (s *Server) GetOrCreatePlayer(playerID string) *models.Player {
	if player, exists := s.gameState.GetPlayer(playerID); exists {
		player.LastSeen = time.Now()
		return player
	}

	// Create new player with default values
	player := models.NewPlayer(playerID)
	s.gameState.SetPlayer(player)
	return player
}

// AddClient registers a new WebSocket client connection with the server.
func (s *Server) AddClient(conn *websocket.Conn, player *models.Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[conn] = player
}

// RemoveClient unregisters a WebSocket client connection from the server.
func (s *Server) RemoveClient(conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.clients, conn)
}

// GetPlayerByConnection retrieves the player associated with a WebSocket connection.
func (s *Server) GetPlayerByConnection(conn *websocket.Conn) *models.Player {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.clients[conn]
}

// GetUpgrader returns the WebSocket upgrader for converting HTTP connections.
func (s *Server) GetUpgrader() *websocket.Upgrader {
	return &s.upgrader
}

// BroadcastToClient sends a message to a specific WebSocket connection.
func (s *Server) BroadcastToClient(conn *websocket.Conn, message []byte) error {
	return conn.WriteMessage(websocket.TextMessage, message)
}
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Game models
type Player struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Factory  *Factory  `json:"factory"`
	Progress *Progress `json:"progress"`
	LastSeen time.Time `json:"lastSeen"`
}

type Factory struct {
	HPStation     *Station `json:"hpStation"`
	ArmorStation  *Station `json:"armorStation"`
	LootStation   *Station `json:"lootStation"`
	AttackStation *Station `json:"attackStation"`
}

type Station struct {
	Level      int     `json:"level"`
	Multiplier float64 `json:"multiplier"`
	Cost       int     `json:"cost"`
}

type Progress struct {
	DungeonLevel int `json:"dungeonLevel"`
	Gold         int `json:"gold"`
	Experience   int `json:"experience"`
}

type Hero struct {
	HP     int `json:"hp"`
	Armor  int `json:"armor"`
	Attack int `json:"attack"`
	Loot   int `json:"loot"`
}

// Game state
type GameState struct {
	Players map[string]*Player `json:"players"`
	mutex   sync.RWMutex
}

type GameServer struct {
	gameState *GameState
	clients   map[*websocket.Conn]*Player
	broadcast chan []byte
	register  chan *websocket.Conn
	upgrader  websocket.Upgrader
}

var gameServer *GameServer

func init() {
	gameServer = &GameServer{
		gameState: &GameState{
			Players: make(map[string]*Player),
		},
		clients:   make(map[*websocket.Conn]*Player),
		broadcast: make(chan []byte),
		register:  make(chan *websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

func main() {
	// Start game loop
	go gameServer.gameLoop()
	go gameServer.handleMessages()

	// Serve static files
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	
	// WebSocket endpoint
	http.HandleFunc("/ws", gameServer.handleWebSocket)
	
	// API endpoints
	http.HandleFunc("/api/player", handlePlayer)
	http.HandleFunc("/api/upgrade", handleUpgrade)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := gs.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	playerID := r.URL.Query().Get("playerID")
	if playerID == "" {
		playerID = generatePlayerID()
	}

	player := gs.getOrCreatePlayer(playerID)
	gs.clients[conn] = player

	// Send initial state
	initialState, _ := json.Marshal(map[string]interface{}{
		"type":   "gameState",
		"player": player,
	})
	conn.WriteMessage(websocket.TextMessage, initialState)

	// Handle messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			delete(gs.clients, conn)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			gs.handleClientMessage(conn, msg)
		}
	}
}

func (gs *GameServer) handleMessages() {
	for {
		select {
		case message := <-gs.broadcast:
			for client := range gs.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("WebSocket write error: %v", err)
					client.Close()
					delete(gs.clients, client)
				}
			}
		}
	}
}

func (gs *GameServer) handleClientMessage(conn *websocket.Conn, msg map[string]interface{}) {
	player := gs.clients[conn]
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
			gs.upgradeStation(player, station)
		}
	}
}

func (gs *GameServer) gameLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gs.gameState.mutex.Lock()
		for _, player := range gs.gameState.Players {
			gs.processPlayer(player)
		}
		gs.gameState.mutex.Unlock()

		// Broadcast updates
		gameUpdate, _ := json.Marshal(map[string]interface{}{
			"type":    "update",
			"players": gs.gameState.Players,
		})
		select {
		case gs.broadcast <- gameUpdate:
		default:
		}
	}
}

func (gs *GameServer) processPlayer(player *Player) {
	// Create hero based on factory stats
	hero := gs.createHero(player.Factory)
	
	// Send hero to battle
	battleResult := gs.battle(hero, player.Progress.DungeonLevel)
	
	// Update player progress
	if battleResult.Victory {
		player.Progress.DungeonLevel++
		player.Progress.Gold += battleResult.GoldReward
		player.Progress.Experience += battleResult.ExpReward
	} else {
		// Still get some reward even on defeat
		player.Progress.Gold += battleResult.GoldReward / 2
	}
}

type BattleResult struct {
	Victory     bool `json:"victory"`
	GoldReward  int  `json:"goldReward"`
	ExpReward   int  `json:"expReward"`
}

func (gs *GameServer) createHero(factory *Factory) *Hero {
	baseHP := 100
	baseArmor := 10
	baseAttack := 20
	baseLoot := 1

	return &Hero{
		HP:     int(float64(baseHP) * factory.HPStation.Multiplier),
		Armor:  int(float64(baseArmor) * factory.ArmorStation.Multiplier),
		Attack: int(float64(baseAttack) * factory.AttackStation.Multiplier),
		Loot:   int(float64(baseLoot) * factory.LootStation.Multiplier),
	}
}

func (gs *GameServer) battle(hero *Hero, dungeonLevel int) BattleResult {
	// Simple battle calculation
	enemyHP := 50 + (dungeonLevel * 10)
	enemyAttack := 15 + (dungeonLevel * 5)
	
	heroHP := hero.HP
	heroDamage := max(1, hero.Attack-enemyAttack/2)
	enemyDamage := max(1, enemyAttack-hero.Armor)
	
	// Battle simulation
	for heroHP > 0 && enemyHP > 0 {
		enemyHP -= heroDamage
		if enemyHP <= 0 {
			break
		}
		heroHP -= enemyDamage
	}
	
	victory := heroHP > 0
	goldReward := (10 + dungeonLevel*2) * hero.Loot
	expReward := 5 + dungeonLevel
	
	return BattleResult{
		Victory:     victory,
		GoldReward:  goldReward,
		ExpReward:   expReward,
	}
}

func (gs *GameServer) getOrCreatePlayer(playerID string) *Player {
	gs.gameState.mutex.Lock()
	defer gs.gameState.mutex.Unlock()

	if player, exists := gs.gameState.Players[playerID]; exists {
		player.LastSeen = time.Now()
		return player
	}

	// Create new player with initial factory
	player := &Player{
		ID:   playerID,
		Name: "Player " + playerID[:8],
		Factory: &Factory{
			HPStation:     &Station{Level: 1, Multiplier: 1.0, Cost: 100},
			ArmorStation:  &Station{Level: 1, Multiplier: 1.0, Cost: 100},
			LootStation:   &Station{Level: 1, Multiplier: 1.0, Cost: 100},
			AttackStation: &Station{Level: 1, Multiplier: 1.0, Cost: 100},
		},
		Progress: &Progress{
			DungeonLevel: 1,
			Gold:         0,
			Experience:   0,
		},
		LastSeen: time.Now(),
	}

	gs.gameState.Players[playerID] = player
	return player
}

func (gs *GameServer) upgradeStation(player *Player, stationType string) {
	var station *Station
	
	switch stationType {
	case "hp":
		station = player.Factory.HPStation
	case "armor":
		station = player.Factory.ArmorStation
	case "loot":
		station = player.Factory.LootStation
	case "attack":
		station = player.Factory.AttackStation
	default:
		return
	}

	if player.Progress.Gold >= station.Cost {
		player.Progress.Gold -= station.Cost
		station.Level++
		station.Multiplier += 0.2 // 20% increase per level
		station.Cost = int(float64(station.Cost) * 1.5) // Cost increases by 50%
	}
}

func handlePlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		http.Error(w, "Player ID required", http.StatusBadRequest)
		return
	}

	player := gameServer.getOrCreatePlayer(playerID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func handleUpgrade(w http.ResponseWriter, r *http.Request) {
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

	player := gameServer.getOrCreatePlayer(playerID)
	gameServer.upgradeStation(player, station)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func generatePlayerID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
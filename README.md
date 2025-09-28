# 🏰 Idle Dungeon

A browser-based idle game written in Go featuring multiplayer elements and a hero factory system inspired by WikiMUD.

## 🎮 Game Overview

In Idle Dungeon, players manage a hero factory that continuously sends heroes into battle to progress through dungeon levels. Each player competes to advance further while upgrading their factory stations for better hero performance.

## 🏗️ Architecture

The project is organized into clean, well-documented packages:

```
idle-dungeon/
├── main.go                 # Server entry point and route setup
├── internal/
│   ├── models/            # Game data structures
│   │   ├── player.go      # Player, Factory, Station, Progress, Hero types
│   │   └── battle.go      # BattleResult type
│   ├── game/              # Core game logic
│   │   ├── server.go      # Game server and multiplayer management
│   │   ├── battle.go      # Combat simulation and hero creation
│   │   └── upgrade.go     # Factory station upgrade logic
│   └── handlers/          # HTTP and WebSocket handlers
│       ├── websocket.go   # Real-time multiplayer communication
│       └── http.go        # REST API endpoints
├── static/                # Frontend assets
│   ├── index.html         # Game web interface
│   ├── style.css          # Responsive styling
│   └── game.js            # Client-side game logic
└── go.mod                 # Go module dependencies
```

## 🏭 Hero Factory System

The core gameplay revolves around four upgradeable factory stations:

- **HP Station**: Increases hero health points (base 100 HP → 1.2x multiplier per upgrade)
- **Armor Station**: Increases hero defense against enemy attacks (base 10 armor)
- **Attack Station**: Increases hero damage output (base 20 attack → 1.2x multiplier per upgrade)
- **Loot Station**: Increases gold rewards from battles (base 1x loot → 1.2x multiplier per upgrade)

Each station starts at level 1 with a 1.0x multiplier and 100 gold cost. Upgrades increase the multiplier by 0.2x and raise the cost by 50% for exponential progression.

## ⚔️ Battle Mechanics

Heroes are automatically generated every second based on current factory station multipliers and sent into battle against dungeon enemies. The battle system uses turn-based combat calculations:

- Enemy difficulty scales with dungeon level (more HP and damage)
- Hero damage is reduced by enemy defense, enemy damage reduced by hero armor
- Victory advances to the next dungeon level and awards full gold/experience
- Defeat still provides partial rewards to maintain progression

## 🌐 Multiplayer Features

Real-time multiplayer is implemented using WebSocket connections:

- Live player list showing all online players sorted by dungeon level
- Real-time updates of player progress, gold, and factory upgrades
- Persistent player state across browser sessions using unique player IDs
- Concurrent game processing for all connected players

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- Modern web browser

### Installation and Running

```bash
# Clone the repository
git clone https://github.com/evevioletrose-hash/idle-dungeon.git
cd idle-dungeon

# Install dependencies
go mod tidy

# Build the server
go build -o idle-dungeon .

# Run the server
./idle-dungeon

# Or run directly with Go
go run main.go
```

The server will start on port 8080 (or the PORT environment variable). Open http://localhost:8080 in your browser to play.

## 🔧 API Endpoints

- `GET /` - Game web interface
- `WS /ws` - WebSocket for real-time multiplayer updates
- `GET /api/player?id={playerID}` - Get player data
- `POST /api/upgrade?playerID={id}&station={type}` - Upgrade factory station

## 📊 Package Documentation

### `internal/models`
Contains all game data structures with comprehensive documentation:
- `Player`: Core player entity with factory and progress
- `Factory`: Hero production facility with four upgrade stations
- `Station`: Individual upgradeable factory components
- `Hero`: Combat units with stats based on factory multipliers
- `GameState`: Thread-safe container for all player data

### `internal/game`
Core game logic and server management:
- `Server`: Multiplayer game server with WebSocket management
- Battle simulation with turn-based combat
- Factory upgrade system with exponential cost scaling
- Concurrent game loop processing all players

### `internal/handlers`
HTTP and WebSocket request handlers:
- WebSocket handler for real-time multiplayer communication
- REST API handlers for player data and upgrades
- Proper error handling and JSON responses

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build and verify
go build .
```

## 🎯 Game Features

- ✅ Idle progression with automatic hero deployment every second
- ✅ Four working factory stations with upgrade mechanics
- ✅ Dungeon level progression with scaling difficulty
- ✅ Real-time multiplayer with WebSocket communication
- ✅ Resource management (gold earning/spending system)
- ✅ Session persistence across browser tabs/refreshes
- ✅ Battle log with timestamped hero deployments and victories
- ✅ Responsive UI that updates in real-time
- ✅ Clean, documented codebase organized into logical packages

## 🤝 Contributing

The codebase is now well-organized and documented, making it easy for humans to read and contribute:

1. All types and functions have comprehensive Go documentation
2. Code is split into logical packages with clear responsibilities
3. Thread-safe operations with proper mutex usage
4. Error handling and logging throughout
5. Consistent naming conventions and code style

Feel free to open issues or submit pull requests!
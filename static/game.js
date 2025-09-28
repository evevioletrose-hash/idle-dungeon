class IdleDungeonGame {
    constructor() {
        this.ws = null;
        this.playerID = this.getOrCreatePlayerID();
        this.player = null;
        this.battleLog = [];
        this.onlinePlayers = {};
        
        this.initializeUI();
        this.connect();
        this.startUpdateLoop();
    }

    getOrCreatePlayerID() {
        let playerID = localStorage.getItem('playerID');
        if (!playerID) {
            playerID = 'player_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('playerID', playerID);
        }
        return playerID;
    }

    initializeUI() {
        document.getElementById('player-id').textContent = this.playerID;
        this.updateConnectionStatus('Connecting...', 'connecting');
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?playerID=${this.playerID}`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            console.log('Connected to game server');
            this.updateConnectionStatus('Connected', 'connected');
        };
        
        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleServerMessage(data);
            } catch (error) {
                console.error('Error parsing server message:', error);
            }
        };
        
        this.ws.onclose = () => {
            console.log('Disconnected from game server');
            this.updateConnectionStatus('Disconnected', 'disconnected');
            // Attempt to reconnect after 3 seconds
            setTimeout(() => this.connect(), 3000);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus('Connection Error', 'disconnected');
        };
    }

    handleServerMessage(data) {
        switch (data.type) {
            case 'gameState':
                this.player = data.player;
                this.updateUI();
                break;
            case 'update':
                if (data.players && data.players[this.playerID]) {
                    const oldLevel = this.player?.progress?.dungeonLevel || 0;
                    this.player = data.players[this.playerID];
                    
                    // Check for level progression
                    if (this.player.progress.dungeonLevel > oldLevel) {
                        this.addBattleLogEntry(`Victory! Advanced to dungeon level ${this.player.progress.dungeonLevel}`, 'victory');
                    }
                }
                this.onlinePlayers = data.players || {};
                this.updateUI();
                break;
        }
    }

    updateUI() {
        if (!this.player) return;

        // Update player info
        document.getElementById('player-name').textContent = this.player.name;

        // Update progress
        document.getElementById('dungeon-level').textContent = this.player.progress.dungeonLevel;
        document.getElementById('gold').textContent = this.player.progress.gold;
        document.getElementById('experience').textContent = this.player.progress.experience;

        // Update factory stations
        this.updateStation('hp', this.player.factory.hpStation);
        this.updateStation('armor', this.player.factory.armorStation);
        this.updateStation('attack', this.player.factory.attackStation);
        this.updateStation('loot', this.player.factory.lootStation);

        // Update online players
        this.updatePlayersList();
    }

    updateStation(stationType, station) {
        document.getElementById(`${stationType}-level`).textContent = station.level;
        document.getElementById(`${stationType}-multiplier`).textContent = station.multiplier.toFixed(1) + 'x';
        document.getElementById(`${stationType}-cost`).textContent = station.cost;

        // Update upgrade button state
        const upgradeBtn = document.querySelector(`[data-station="${stationType}"] .upgrade-btn`);
        if (this.player.progress.gold >= station.cost) {
            upgradeBtn.disabled = false;
            upgradeBtn.textContent = 'Upgrade';
        } else {
            upgradeBtn.disabled = true;
            upgradeBtn.textContent = `Need ${station.cost - this.player.progress.gold} more gold`;
        }
    }

    updatePlayersList() {
        const playersContainer = document.getElementById('online-players');
        playersContainer.innerHTML = '';

        const playerList = Object.values(this.onlinePlayers).sort((a, b) => {
            return b.progress.dungeonLevel - a.progress.dungeonLevel;
        });

        playerList.forEach((player, index) => {
            const playerElement = document.createElement('div');
            playerElement.className = 'player-item';
            if (player.id === this.playerID) {
                playerElement.classList.add('current-player');
            }

            playerElement.innerHTML = `
                <div class="player-name">${player.name}</div>
                <div class="player-stats">
                    Level: ${player.progress.dungeonLevel}<br>
                    Gold: ${player.progress.gold}
                </div>
            `;

            playersContainer.appendChild(playerElement);
        });

        if (playerList.length === 0) {
            playersContainer.innerHTML = '<div class="loading">No players online</div>';
        }
    }

    updateConnectionStatus(status, className) {
        const statusElement = document.getElementById('connection-status');
        statusElement.textContent = status;
        statusElement.className = className;
    }

    addBattleLogEntry(message, type = '') {
        const logContainer = document.getElementById('battle-log');
        const entry = document.createElement('div');
        entry.className = `log-entry ${type}`;
        entry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
        
        logContainer.appendChild(entry);
        logContainer.scrollTop = logContainer.scrollHeight;

        // Keep only last 50 entries
        while (logContainer.children.length > 50) {
            logContainer.removeChild(logContainer.firstChild);
        }
    }

    startUpdateLoop() {
        // Update battle log periodically with current hero stats
        setInterval(() => {
            if (this.player) {
                const hero = this.calculateCurrentHero();
                const message = `Hero deployed: HP:${hero.hp} ATK:${hero.attack} ARM:${hero.armor} LOOT:${hero.loot}`;
                this.addBattleLogEntry(message);
            }
        }, 5000);
    }

    calculateCurrentHero() {
        if (!this.player) return { hp: 0, attack: 0, armor: 0, loot: 0 };

        const factory = this.player.factory;
        return {
            hp: Math.floor(100 * factory.hpStation.multiplier),
            attack: Math.floor(20 * factory.attackStation.multiplier),
            armor: Math.floor(10 * factory.armorStation.multiplier),
            loot: Math.floor(1 * factory.lootStation.multiplier)
        };
    }

    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }
}

// Global functions for HTML onclick handlers
function upgradeStation(stationType) {
    if (window.game) {
        window.game.sendMessage({
            type: 'upgrade',
            station: stationType
        });
    }
}

// Initialize game when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.game = new IdleDungeonGame();
});

// Handle page visibility changes to maintain connection
document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible' && window.game) {
        // Reconnect if disconnected when page becomes visible
        if (!window.game.ws || window.game.ws.readyState !== WebSocket.OPEN) {
            window.game.connect();
        }
    }
});
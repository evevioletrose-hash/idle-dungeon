package game

import (
	"github.com/evevioletrose-hash/idle-dungeon/internal/models"
)

// processPlayer handles the battle logic for a single player.
// It creates a hero based on factory stats, simulates battle, and updates progress.
func (s *Server) processPlayer(player *models.Player) {
	// Create hero based on current factory station multipliers
	hero := s.createHero(player.Factory)
	
	// Simulate battle against dungeon enemy
	battleResult := s.simulateBattle(hero, player.Progress.DungeonLevel)
	
	// Update player progress based on battle outcome
	if battleResult.Victory {
		player.Progress.DungeonLevel++
		player.Progress.Gold += battleResult.GoldReward
		player.Progress.Experience += battleResult.ExpReward
	} else {
		// Partial rewards even on defeat to maintain progression
		player.Progress.Gold += battleResult.GoldReward / 2
	}
}

// createHero generates a hero with stats based on factory station multipliers.
// Base stats are modified by each station's current multiplier value.
func (s *Server) createHero(factory *models.Factory) *models.Hero {
	// Base hero statistics
	const (
		baseHP     = 100
		baseArmor  = 10
		baseAttack = 20
		baseLoot   = 1
	)

	return &models.Hero{
		HP:     int(float64(baseHP) * factory.HPStation.Multiplier),
		Armor:  int(float64(baseArmor) * factory.ArmorStation.Multiplier),
		Attack: int(float64(baseAttack) * factory.AttackStation.Multiplier),
		Loot:   int(float64(baseLoot) * factory.LootStation.Multiplier),
	}
}

// simulateBattle performs turn-based combat between a hero and dungeon enemy.
// Enemy difficulty scales with dungeon level, and rewards are based on enemy strength.
func (s *Server) simulateBattle(hero *models.Hero, dungeonLevel int) models.BattleResult {
	// Enemy stats scale with dungeon level
	enemyHP := 50 + (dungeonLevel * 10)
	enemyAttack := 15 + (dungeonLevel * 5)
	
	// Combat variables
	heroHP := hero.HP
	heroDamage := max(1, hero.Attack-enemyAttack/2) // Hero damage reduced by enemy attack/2
	enemyDamage := max(1, enemyAttack-hero.Armor)   // Enemy damage reduced by hero armor
	
	// Turn-based battle simulation
	for heroHP > 0 && enemyHP > 0 {
		// Hero attacks first
		enemyHP -= heroDamage
		if enemyHP <= 0 {
			break // Hero wins
		}
		
		// Enemy counter-attacks
		heroHP -= enemyDamage
	}
	
	// Determine battle outcome and calculate rewards
	victory := heroHP > 0
	goldReward := (10 + dungeonLevel*2) * hero.Loot // Gold scales with level and loot multiplier
	expReward := 5 + dungeonLevel                   // Experience scales with dungeon level
	
	return models.BattleResult{
		Victory:     victory,
		GoldReward:  goldReward,
		ExpReward:   expReward,
	}
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
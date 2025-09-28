package models

// BattleResult represents the outcome of a hero's battle against a dungeon enemy.
// It contains information about victory/defeat and rewards earned.
type BattleResult struct {
	Victory     bool `json:"victory"`     // Whether the hero won the battle
	GoldReward  int  `json:"goldReward"`  // Gold earned from the battle
	ExpReward   int  `json:"expReward"`   // Experience points earned from the battle
}
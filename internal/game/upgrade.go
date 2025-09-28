package game

import (
	"github.com/evevioletrose-hash/idle-dungeon/internal/models"
)

// UpgradeStation attempts to upgrade a specific factory station for a player.
// It checks if the player has enough gold, then increases the station's level,
// multiplier, and cost according to the game's progression rules.
func (s *Server) UpgradeStation(player *models.Player, stationType string) bool {
	station := s.getStationByType(player.Factory, stationType)
	if station == nil {
		return false // Invalid station type
	}

	// Check if player has enough gold for the upgrade
	if player.Progress.Gold < station.Cost {
		return false // Insufficient funds
	}

	// Perform the upgrade
	player.Progress.Gold -= station.Cost          // Deduct upgrade cost
	station.Level++                               // Increase station level
	station.Multiplier += 0.2                    // Increase effectiveness by 20%
	station.Cost = int(float64(station.Cost) * 1.5) // Increase next upgrade cost by 50%

	return true // Upgrade successful
}

// getStationByType returns the appropriate station pointer based on the station type string.
// This is a helper function to map string identifiers to actual station objects.
func (s *Server) getStationByType(factory *models.Factory, stationType string) *models.Station {
	switch stationType {
	case "hp":
		return factory.HPStation
	case "armor":
		return factory.ArmorStation
	case "loot":
		return factory.LootStation
	case "attack":
		return factory.AttackStation
	default:
		return nil // Unknown station type
	}
}
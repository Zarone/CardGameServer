package gamemanager_test

import (
	"testing"

	"github.com/Zarone/CardGameServer/cmd/gamemanager"
)

func TestMakeGame(t *testing.T) {
	game := gamemanager.MakeGame()
	
	if game == nil {
		t.Error("MakeGame returned nil")
	}
	
	if game.CardIndex != 0 {
		t.Errorf("Expected CardIndex 0, got %d", game.CardIndex)
	}
	
	if game.Players == nil {
		t.Error("Game players slice is nil")
	}
	
	if len(game.Players) != 0 {
		t.Errorf("Expected empty players slice, got length %d", len(game.Players))
	}
}

func TestGameAddPlayer(t *testing.T) {
	game := gamemanager.MakeGame()
	
	// Test adding first player
	playerID := game.AddPlayer()
	if playerID != 0 {
		t.Errorf("Expected player ID 0, got %d", playerID)
	}
	
	if len(game.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(game.Players))
	}
	
	// Test adding second player
	playerID = game.AddPlayer()
	if playerID != 1 {
		t.Errorf("Expected player ID 1, got %d", playerID)
	}
	
	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(game.Players))
	}
}

func TestGameSetupPlayer(t *testing.T) {
	game := gamemanager.MakeGame()
	playerID := game.AddPlayer()
	
	// Test setting up player with empty deck
	deck := []uint{}
	gameIDs := game.SetupPlayer(playerID, deck)
	
	if gameIDs == nil {
		t.Error("SetupPlayer returned nil")
		return
	}
	
	if len(*gameIDs) != 0 {
		t.Errorf("Expected empty game IDs slice, got length %d", len(*gameIDs))
	}
	
	// Test setting up player with some cards
	deck = []uint{1, 2, 3, 4, 5}
	gameIDs = game.SetupPlayer(playerID, deck)
	
	if gameIDs == nil {
		t.Error("SetupPlayer returned nil")
		return
	}
	
	if len(*gameIDs) != len(deck) {
		t.Errorf("Expected %d game IDs, got %d", len(deck), len(*gameIDs))
	}
}

func TestGameStartGame(t *testing.T) {
	game := gamemanager.MakeGame()
	
	// Add two players
	game.AddPlayer()
	game.AddPlayer()
	
	// Set up players with some cards
	deck := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	game.SetupPlayer(0, deck)
	game.SetupPlayer(1, deck)
	
	// Start the game
	p1Moves, p2Moves := game.StartGame()
	
	if p1Moves == nil || p2Moves == nil {
		t.Error("StartGame returned nil moves")
		return
	}
	
	if len(*p1Moves) != 7 || len(*p2Moves) != 7 {
		t.Errorf("Expected 7 moves per player, got %d and %d", len(*p1Moves), len(*p2Moves))
	}

	// Check that all cards moved from deck to hand
	for _, move := range *p1Moves {
		if move.From != gamemanager.DECK_PILE || move.To != gamemanager.HAND_PILE {
			t.Errorf("Expected move from DECK to HAND, got from %s to %s", move.From, move.To)
		}
	}

	// Check for duplicates in hand
	seen := make(map[uint]bool)
	for _, move := range *p1Moves {
		if seen[move.CardID] {
			t.Errorf("Duplicate card ID %d found in hand", move.CardID)
		}
		seen[move.CardID] = true
	}
} 
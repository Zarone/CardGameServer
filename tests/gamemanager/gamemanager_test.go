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
			t.Errorf("Expected move from DECK to HAND, got from %v to %v", move.From, move.To)
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

	// Verify that cards in hand are not in deck
	player1 := game.Players[0]
	for _, card := range player1.Hand.Cards {
		// Check if card is still in deck
		for _, deckCard := range player1.Deck.Cards {
			if card.GameID == deckCard.GameID {
				t.Errorf("Card with GameID %d found in both hand and deck", card.GameID)
			}
		}
	}

	// Verify that moves accurately reflect the cards in hand
	moveCardIDs := make(map[uint]bool)
	for _, move := range *p1Moves {
		moveCardIDs[move.CardID] = true
	}

	for _, card := range player1.Hand.Cards {
		if !moveCardIDs[card.GameID] {
			t.Errorf("Card with GameID %d in hand but not listed in moves", card.GameID)
		}
	}

	for cardID := range moveCardIDs {
		found := false
		for _, card := range player1.Hand.Cards {
			if card.GameID == cardID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Move lists card ID %d but card not found in hand", cardID)
		}
	}
}

func TestGameStartGameWithFewCards(t *testing.T) {
	game := gamemanager.MakeGame()
	
	// Add two players
	game.AddPlayer()
	game.AddPlayer()
	
	// Set up players with fewer than 7 cards
	deck := []uint{1, 2, 3}
	game.SetupPlayer(0, deck)
	game.SetupPlayer(1, deck)
	
	// Start the game
	p1Moves, p2Moves := game.StartGame()
	
	if p1Moves == nil || p2Moves == nil {
		t.Error("StartGame returned nil moves")
		return
	}
	
	// Should get all available cards (3) instead of 7
	if len(*p1Moves) != 3 || len(*p2Moves) != 3 {
		t.Errorf("Expected 3 moves per player (all available cards), got %d and %d", len(*p1Moves), len(*p2Moves))
	}

	// Check that all cards moved from deck to hand
	for _, move := range *p1Moves {
		if move.From != gamemanager.DECK_PILE || move.To != gamemanager.HAND_PILE {
			t.Errorf("Expected move from DECK to HAND, got from %v to %v", move.From, move.To)
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

	// Verify that cards in hand are not in deck
	player1 := game.Players[0]
	for _, card := range player1.Hand.Cards {
		// Check if card is still in deck
		for _, deckCard := range player1.Deck.Cards {
			if card.GameID == deckCard.GameID {
				t.Errorf("Card with GameID %d found in both hand and deck", card.GameID)
			}
		}
	}

	// Verify that moves accurately reflect the cards in hand
	moveCardIDs := make(map[uint]bool)
	for _, move := range *p1Moves {
		moveCardIDs[move.CardID] = true
	}

	for _, card := range player1.Hand.Cards {
		if !moveCardIDs[card.GameID] {
			t.Errorf("Card with GameID %d in hand but not listed in moves", card.GameID)
		}
	}

	for cardID := range moveCardIDs {
		found := false
		for _, card := range player1.Hand.Cards {
			if card.GameID == cardID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Move lists card ID %d but card not found in hand", cardID)
		}
	}

	// Verify deck is empty after moving all cards
	if len(player1.Deck.Cards) != 0 {
		t.Errorf("Expected empty deck after moving all cards, got %d cards", len(player1.Deck.Cards))
	}
} 

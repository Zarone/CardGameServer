package gamemanager_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Zarone/CardGameServer/cmd/gamemanager"
)

var cardInfoPath string = "../../cardInfo"

func TestMakeGame(t *testing.T) {
	game := gamemanager.MakeGame(gamemanager.SetupFromDirectory(cardInfoPath))
	
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
	game := gamemanager.MakeGame(gamemanager.SetupFromDirectory(cardInfoPath))
	
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
	game := gamemanager.MakeGame(gamemanager.SetupFromDirectory(cardInfoPath))
	
	// Add two players
	game.AddPlayer()
	game.AddPlayer()
	
	// Set up players with some cards
	deck := []uint{1, 1, 1, 1, 1, 2, 2, 2, 2, 2}
	game.SetupPlayer(0, deck)
	game.SetupPlayer(1, deck)
	
	// Start the game
	p1Info, p2Info := game.StartGame(true)
	p1Moves, p2Moves := &p1Info.Movements, &p2Info.Movements
	
	if len(*p1Moves) != 14 || len(*p2Moves) != 14 {
		t.Errorf("Expected 14 moves per player, got %d and %d", len(*p1Moves), len(*p2Moves))
	}

	// Check that all cards moved from deck to hand
	for _, move := range *p1Moves {
    if !( 
      (move.From == gamemanager.DECK_PILE && move.To == gamemanager.HAND_PILE) ||
      (move.From == gamemanager.OPP_DECK_PILE && move.To == gamemanager.OPP_HAND_PILE) ) {
			t.Errorf("Expected move from DECK to HAND, got from %v to %v", move.From, move.To)
    }
	}

	// Check for duplicates in hand
	seen := make(map[uint]bool)
	for _, move := range *p1Moves {
		if seen[move.GameID] {
			t.Errorf("Duplicate card ID %d found in hand", move.GameID)
		}
		seen[move.GameID] = true
	}

	// Verify that cards in hand are not in deck
	player1 := game.Players[0]
	for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
		// Check if card is still in deck
		for _, deckCard := range player1.PlayerPiles[gamemanager.DECK_PILE].Cards {
			if card.GameID == deckCard.GameID {
				t.Errorf("Card with GameID %d found in both hand and deck", card.GameID)
			}
		}
	}

	// Verify that moves accurately reflect the cards in hand
	moveGameIDs := make(map[uint]bool)
	for _, move := range *p1Moves {
    if move.From == gamemanager.OPP_DECK_PILE { continue }
		moveGameIDs[move.GameID] = true
	}

	for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
		fmt.Println(card, "in hand")
		if !moveGameIDs[card.GameID] {
			t.Errorf("Card with GameID %d in hand but not listed in moves", card.GameID)
		}
	}

	for cardID := range moveGameIDs {
		found := false
		for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
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
	game := gamemanager.MakeGame(gamemanager.SetupFromDirectory(cardInfoPath))
	
	// Add two players
	game.AddPlayer()
	game.AddPlayer()
	
	// Set up players with fewer than 7 cards
	deck := []uint{1, 2, 3}
	game.SetupPlayer(0, deck)
	game.SetupPlayer(1, deck)
	
	// Start the game
	p1Info, p2Info := game.StartGame(true)
	p1Moves, p2Moves := &p1Info.Movements, &p2Info.Movements
	
	// Should get all available cards (3) instead of 7
	if len(*p1Moves) != 6 || len(*p2Moves) != 6 {
		t.Errorf("Expected 3 moves per player (all available cards), got %d and %d", len(*p1Moves), len(*p2Moves))
	}

	// Check that all cards moved from deck to hand
	for _, move := range *p1Moves {
    if !( 
      (move.From == gamemanager.DECK_PILE && move.To == gamemanager.HAND_PILE) ||
      (move.From == gamemanager.OPP_DECK_PILE && move.To == gamemanager.OPP_HAND_PILE) ) {
			t.Errorf("Expected move from DECK to HAND, got from %v to %v", move.From, move.To)
		}
	}

	// Check for duplicates in hand
	seen := make(map[uint]bool)
	for _, move := range *p1Moves {
		if seen[move.GameID] {
			t.Errorf("Duplicate card ID %d found in hand", move.GameID)
		}
		seen[move.GameID] = true
	}

	// Verify that cards in hand are not in deck
	player1 := game.Players[0]
	for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
		// Check if card is still in deck
		for _, deckCard := range player1.PlayerPiles[gamemanager.DECK_PILE].Cards {
			if card.GameID == deckCard.GameID {
				t.Errorf("Card with GameID %d found in both hand and deck", card.GameID)
			}
		}
	}

	// Verify that moves accurately reflect the cards in hand
	moveGameIDs := make(map[uint]bool)
	for _, move := range *p1Moves {
    if move.From == gamemanager.OPP_DECK_PILE { continue }
		moveGameIDs[move.GameID] = true
	}

	for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
		if !moveGameIDs[card.GameID] {
			t.Errorf("Card with GameID %d in hand but not listed in moves", card.GameID)
		}
	}

	for gameID := range moveGameIDs {
		found := false
		for _, card := range player1.PlayerPiles[gamemanager.HAND_PILE].Cards {
			if card.GameID == gameID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Move lists card ID %d but card not found in hand", gameID)
		}
	}

	// Verify deck is empty after moving all cards
	if len(player1.PlayerPiles[gamemanager.DECK_PILE].Cards) != 0 {
		t.Errorf("Expected empty deck after moving all cards, got %d cards", len(player1.PlayerPiles[gamemanager.DECK_PILE].Cards))
	}
} 

func TestPlayUltraBall(t *testing.T) {
	game := gamemanager.MakeGame(gamemanager.SetupFromString(`[
    {
      "name": "card 5",
      "imageSrc": "card5"
    },
    {
      "name": "Ultra Ball",
      "imageSrc": "card6",
      "preCondition": {
        "kind": "OPERATOR",
        "operator": ">",
        "args": [
          {
            "kind": "VARIABLE",
            "variable": "CARDS_IN_HAND"
          },
          {
            "kind": "CONSTANT",
            "val": 2
          }
        ]
      },
      "effect": { 
        "kind": "THEN", 
        "args": [
          {
            "kind": "MOVE",
            "target": { "kind": "TARGET", "targetType": "THIS" },
            "to": "DISCARD"
          },
          {
            "kind": "MOVE",
            "target": {
              "kind": "TARGET",
              "targetType": "SELECT",
              "filter": {
                "kind": "JUST",
                "pile": "HAND",
                "count": {
                  "atLeast": 2,
                  "atMost": 2
                }
              }
            },
            "to": "DISCARD"
          }
        ]
      }
    }
  ]`))
	
	// Add two players
	game.AddPlayer()
	game.AddPlayer()
	
	deck := []uint{0, 0, 1}
	game.SetupPlayer(0, deck)
	game.SetupPlayer(1, deck)
	
	// Start the game
  p1Start, _ := game.StartGame(true)
  strData, err := json.Marshal(p1Start)
  if err != nil {
    fmt.Println("Error getting json", err)
  }
  fmt.Println("start info", string(strData))

  if len(p1Start.SelectableCards) != 3 {
    t.Error("Initial selectable cards not correct")
  }

  info, oppInfo, err := game.ProcessAction(0, &gamemanager.Action{
    ActionType: gamemanager.ActionTypeSelectCard,
    SelectedCards: append(make([]uint, 0, 1), 2),
    From: gamemanager.HAND_PILE,
  })
  if err != nil {
    fmt.Println("Error processing action:", err)
  }

  strData, err = json.Marshal(info)
  if err != nil {
    fmt.Println("Error getting json", err)
  }
  fmt.Println("resulting info", string(strData))

  if oppInfo == nil {
    t.Error("no opponent info")
  }

  // should be selecting cards after played ultra ball
  if info.Phase != gamemanager.PHASE_SELECTING_CARDS {
    t.Errorf("Not set to selecting cards")
  }

  if len(info.Movements) != 1 || 
    info.Movements[0].From != gamemanager.HAND_PILE || 
    info.Movements[0].To != gamemanager.DISCARD_PILE || 
    info.Movements[0].GameID != 2 {
    
    t.Error("Does not send ultra ball to discard")
  }

  if info.Pile != gamemanager.HAND_PILE {
    t.Error("Does not keep hand open")
  }

  if len(info.OpenViewCards) != 0 {
    t.Error("open view cards is non-empty")
  }

  if len(info.SelectableCards) != 2 || info.SelectableCards[0] == 2 || info.SelectableCards[1] == 2 {
    t.Error("cards in hand aren't correctly selectable")
  }

  if len(game.Players[0].PlayerPiles[gamemanager.HAND_PILE].Cards) != 2 || 
  len(game.Players[0].PlayerPiles[gamemanager.DISCARD_PILE].Cards) != 1 {
    t.Error("didn't actually discard ultra ball")
  }

  info, oppInfo, err = game.ProcessAction(0, &gamemanager.Action{
    ActionType: gamemanager.ActionTypeFinishSelection,
    SelectedCards: append(make([]uint, 0, 2), 0, 1),
    From: gamemanager.HAND_PILE,
  })

  if oppInfo == nil {
    t.Error("No opponent info")
  }

  if err != nil {
    fmt.Println("Error processing action:", err)
  }

  strData, err = json.Marshal(info)
  if err != nil {
    fmt.Println("Error getting json", err)
  }
  fmt.Println("resulting info", string(strData))

  if info.Phase != gamemanager.PHASE_MY_TURN {
    t.Errorf("Not set back to my turn")
  }

  if len(info.Movements) != 2 {
    t.Error("Does not send cards to discard")
  }
  // assert they move to discard
  for _, move := range info.Movements {
    if move.From != gamemanager.HAND_PILE || move.To != gamemanager.DISCARD_PILE {
      t.Error("Invalid move")
    }
  }

  if info.Pile != gamemanager.HAND_PILE {
    t.Error("Does not keep hand open")
  }

  if len(info.OpenViewCards) != 0 {
    t.Error("open view cards is non-empty")
  }

  if len(info.SelectableCards) != 0 {
    t.Error("cards in hand aren't correctly selectable")
  }

  if len(game.Players[0].PlayerPiles[gamemanager.HAND_PILE].Cards) != 0 || 
  len(game.Players[0].PlayerPiles[gamemanager.DISCARD_PILE].Cards) != 3 {
    t.Error("didn't actually discard ultra ball")
  }

}

package gamemanager

import "fmt"

type Action struct {
  ActionType    ActionType  `json:"type"`
  SelectedCards []uint      `json:"selectedCards"`
  From          Pile        `json:"from"`
}

func (a *Action) String() string {
  return fmt.Sprintf("{ActionType: %v, SelectedCards: %v, From: %v}\n", a.ActionType, a.SelectedCards, a.From)
}

type Game struct {
	Players    []Player
	CardIndex  uint
}

// returns the index of this player within
// the players of this game
func (g *Game) AddPlayer() uint8 {
	var newIndex uint8 = uint8(len(g.Players))
	g.Players = append(g.Players, Player{
		Deck: CardGroup{
			Cards: make([]Card, 0),
			Pile: DECK_PILE,
		},
		Hand: CardGroup{
			Cards: make([]Card, 0),
			Pile: HAND_PILE,
		},
	})
	return newIndex
}

// Sets up player with the given playerID with the deck given by 
// an array of the card IDs.
func (g *Game) SetupPlayer(playerID uint8, deck []uint) {
	var player *Player = &g.Players[playerID]
	player.Deck.Cards = make([]Card, 0, len(deck))
	for _, el := range deck {
		player.Deck.Cards = append(player.Deck.Cards, Card{
			ID: el,
			GameID: g.CardIndex,
		})
		g.CardIndex++
	}
}

func (g *Game) GetSetupData(playerID uint8) (*[]uint, *[]uint) {
  myDeck := make([]uint, 0, len(g.Players[playerID].Deck.Cards))
	for _, el := range g.Players[playerID].Deck.Cards {
    myDeck = append(myDeck, el.GameID)
	}

  oppDeck := make([]uint, 0, len(g.Players[1-playerID].Deck.Cards))
	for _, el := range g.Players[1-playerID].Deck.Cards {
    oppDeck = append(oppDeck, el.GameID)
	}

	return &myDeck, &oppDeck
}

func (g *Game) String() string {
	str := "[Game:\n"

	for _, el := range g.Players {
		str += el.String() + ",\n"
	}

	return str + "]"
}

func (g *Game) StartGame() (*[]CardMovement, *[]CardMovement) {
	g.Players[0].Deck.shuffle()
	g.Players[1].Deck.shuffle()
  fmt.Println(g.Players[0], g.Players[1])
  out1, out2 := g.Players[0].Deck.moveFromTopTo(&g.Players[0].Hand, 7), 
		g.Players[1].Deck.moveFromTopTo(&g.Players[1].Hand, 7)
  fmt.Println(g.Players[0], g.Players[1])
	return out1, out2
}

func (g *Game) ProcessAction(user uint8, action *Action) (UpdateInfo, error) {

  if (ActionType(action.ActionType) == ActionTypeSelectCard) {
    fmt.Printf("Action: Play Card\n")

    if (len(action.SelectedCards) != 1) {
      return UpdateInfo{}, fmt.Errorf("Play card was triggered with multiple cards")
    }

    if (action.From == HAND_PILE) {
      card := g.Players[user].Hand.find(action.SelectedCards[0])
      if card == nil {
        fmt.Printf("Can't find card\n")
        return UpdateInfo{}, fmt.Errorf("Can't find card\n")
      }

      if card.ID == 5 {
        movements := make([]CardMovement, 0, 1)
        movements = append(movements, CardMovement{
          From: HAND_PILE,
          To: DISCARD_PILE,
          GameID: card.GameID,
          CardID: card.ID,
        })
        selectableCards := make([]uint, 0, len(g.Players[user].Hand.Cards)-1)
        for _, thisCard := range g.Players[user].Hand.Cards {
          if thisCard.GameID != card.GameID {
            selectableCards = append(selectableCards, thisCard.GameID)
          }
        }
        return UpdateInfo{
          Movements: movements,
          Phase: PHASE_SELECTING_CARDS,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: selectableCards,
        }, nil
      } else {
        selectableCards := make([]uint, 0, len(g.Players[user].Hand.Cards))
        for _, thisCard := range g.Players[user].Hand.Cards {
          selectableCards = append(selectableCards, thisCard.GameID)
        }
        return UpdateInfo{
          Movements: append(make([]CardMovement, 0, 1), CardMovement{
            From: HAND_PILE,
            To: DISCARD_PILE,
            GameID: action.SelectedCards[0],
            CardID: card.ID,
          }),
          Phase: PHASE_MY_TURN,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: selectableCards,
        }, nil
      }
    }
  } else if ActionType(action.ActionType) == ActionTypeFinishSelection {
    movements := make([]CardMovement, 0, len(action.SelectedCards))
    for _, el := range action.SelectedCards {
      movements = append(movements, CardMovement{
        From: HAND_PILE,
        To: DISCARD_PILE,
        GameID: el,
        CardID: g.Players[user].Hand.find(el).ID, 
      })
    }
    selectableCards := make([]uint, 0, len(g.Players[user].Hand.Cards))
    for _, thisCard := range g.Players[user].Hand.Cards {
      selectableCards = append(selectableCards, thisCard.GameID)
    }
    return UpdateInfo{
      Movements: movements,
      Phase: PHASE_MY_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: selectableCards,
    }, nil
  }

  return UpdateInfo{}, fmt.Errorf("Not sure how to handle action")
}

func MakeGame() *Game {
	return &Game{
		CardIndex: 0,
		Players: make([]Player, 0, 2),
	}
}



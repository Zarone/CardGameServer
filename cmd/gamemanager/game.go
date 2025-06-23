package gamemanager

import "fmt"

type Game struct {
	Players         []Player
	CardIndex       uint
  CardHandler     *CardHandler
  CardActionStack *CardActionStack
}

func MakeGame(cardHandler *CardHandler) *Game {
	return &Game{
		CardIndex: 0,
		Players: make([]Player, 0, 2),
    CardHandler: cardHandler,
    CardActionStack: nil,
	}
}

// returns the index of this player within
// the players of this game
func (g *Game) AddPlayer() uint8 {
	var newIndex uint8 = uint8(len(g.Players))
	g.Players = append(g.Players, MakePlayer())
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
    player.FindID[g.CardIndex] = &player.Deck
		g.CardIndex++
	}
}

// Takes cardIDs, and returns the corresponding game IDs
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

func (g *Game) StartGame(goingFirst bool) (*UpdateInfo, *UpdateInfo) {
	g.Players[0].Deck.shuffle()
	g.Players[1].Deck.shuffle()
  fmt.Println(g.Players[0], g.Players[1])
  p1Moves, p2Moves := g.Players[0].moveFromTopTo(&g.Players[0].Deck, &g.Players[0].Hand, 7), 
		g.Players[1].moveFromTopTo(&g.Players[1].Deck, &g.Players[1].Hand, 7)
  fmt.Println(g.Players[0], g.Players[1])

  var selectableCards []uint
  var phase Phase
  if goingFirst {
    phase = PHASE_MY_TURN
    selectableCards = *g.getPlayableCards(0)
  } else {
    phase = PHASE_OPPONENTS_TURN
    selectableCards = make([]uint, 0)
  }

  out1 := UpdateInfo{
    Movements: *mergeMoves(p1Moves, p2Moves),
    Phase: phase,
    Pile: HAND_PILE,
    OpenViewCards: make([]uint, 0),
    SelectableCards: selectableCards,
  }

  if goingFirst {
    phase = PHASE_OPPONENTS_TURN
    selectableCards = make([]uint, 0)
  } else {
    phase = PHASE_MY_TURN
    selectableCards = *g.getPlayableCards(1) 
  }
  out2 := UpdateInfo{
    Movements: *mergeMoves(p2Moves, p1Moves),
    Phase: phase,
    Pile: HAND_PILE,
    OpenViewCards: make([]uint, 0),
    SelectableCards: selectableCards,
  }

	return &out1, &out2
}

func (g *Game) ProcessAction(user uint8, action *Action) (*UpdateInfo, *UpdateInfo, error) {
  if (ActionType(action.ActionType) == ActionTypeSelectCard) {
    fmt.Printf("Action: Play Card\n")

    if (len(action.SelectedCards) != 1) {
      return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("play card was triggered with multiple cards")
    }

    if (action.From == HAND_PILE) {
      card := g.Players[user].Hand.find(action.SelectedCards[0])
      if card == nil {
        fmt.Printf("Can't find card\n")
        return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Can't find card\n")
      }

      staticCardData := g.CardHandler.cardLookup["set1"][card.ID]

      if staticCardData.Effect != nil { 
        g.CardActionStack = nil
        info, _, err := g.processCardAction(user, staticCardData.Effect, action, nil)
        if err != nil {
          return nil, nil, err
        }
        return info, toOppInfo(info), nil
      } else {
        movements := append(
          make([]CardMovement, 0, 1), 
          g.Players[user].moveCardTo(
            action.SelectedCards[0], 
            &g.Players[user].Discard,
          ),
        )
        info := &UpdateInfo{
          Movements: movements,
          Phase: PHASE_MY_TURN,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: *g.getPlayableCards(user),
        }
        return info, toOppInfo(info), nil
      }
    }
  } else if ActionType(action.ActionType) == ActionTypeFinishSelection {
    if g.CardActionStack != nil {
      info, _, err := g.processCardAction(user, nil, action, nil)
      if err != nil {
        return nil, nil, err
      }
      return info, toOppInfo(info), nil
    }

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

    info := &UpdateInfo{
      Movements: movements,
      Phase: PHASE_MY_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: selectableCards,
    }
    return info, toOppInfo(info), nil
  }

  return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Not sure how to handle action")
}

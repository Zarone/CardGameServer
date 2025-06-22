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
	Players         []Player
	CardIndex       uint
  CardHandler     *CardHandler
  CardActionStack *CardActionStack
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

// Takes a pile and converts it to the opponent's equivalent of that pile
func toOpp(pile Pile) Pile {
  if pile == HAND_PILE {
    return OPP_HAND_PILE
  } else if pile == RESERVE_PILE {
    return OPP_RESERVE_PILE
  } else if pile == SPECIAL_PILE {
    return OPP_SPECIALS_PILE
  } else if pile == BATTLEFIELD_PILE {
    return OPP_BATTLEFIELD_PILE
  } else if pile == DISCARD_PILE {
    return OPP_DISCARD_PILE
  } else if pile == DECK_PILE {
    return OPP_DECK_PILE
  } else {
    fmt.Println("Not sure what opponent's version of this pile is", pile)
    return pile
  }
}


// Takes moves of player and opponent and returns a merged cardMovement slice to send to player
func mergeMoves(thisPlayerMoves *[]CardMovement, oppPlayerMoves *[]CardMovement) *[]CardMovement {
  ret := make([]CardMovement, 0, len(*thisPlayerMoves)+len(*oppPlayerMoves))
  for _, movement := range *thisPlayerMoves {
    ret = append(ret, movement)
  }
  for _, movement := range *oppPlayerMoves{
    ret = append(ret, CardMovement{
      From: toOpp(movement.From),
      To: toOpp(movement.To),
      GameID: movement.GameID,
      CardID: movement.CardID,
    })
  }
  return &ret
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
    selectableCards = make([]uint, 0, len(g.Players[0].Hand.Cards))
    for _, thisCard := range g.Players[0].Hand.Cards {
      selectableCards = append(selectableCards, thisCard.GameID)
    }
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
    selectableCards = make([]uint, 0, len(g.Players[1].Hand.Cards))
    for _, thisCard := range g.Players[1].Hand.Cards {
      selectableCards = append(selectableCards, thisCard.GameID)
    }
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
  if g.CardActionStack != nil {
    g.processCardAction(user, nil, action, nil)
  }

  if (ActionType(action.ActionType) == ActionTypeSelectCard) {
    fmt.Printf("Action: Play Card\n")

    if (len(action.SelectedCards) != 1) {
      return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Play card was triggered with multiple cards")
    }

    if (action.From == HAND_PILE) {
      card := g.Players[user].Hand.find(action.SelectedCards[0])
      if card == nil {
        fmt.Printf("Can't find card\n")
        return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Can't find card\n")
      }

      fmt.Println(g.CardHandler.cardLookup["set1"][card.ID])
      staticCardData := g.CardHandler.cardLookup["set1"][card.ID]

      if staticCardData.Effect != nil { 
        g.CardActionStack = nil
        info, _, err := g.processCardAction(user, staticCardData.Effect, action, nil)
        if err != nil {
          return nil, nil, err
        }
        return info, nil, nil
      //if card.ID == 5 {
        //movements := make([]CardMovement, 0, 1)
        //movements = append(movements, CardMovement{
          //From: HAND_PILE,
          //To: DISCARD_PILE,
          //GameID: card.GameID,
          //CardID: card.ID,
        //})
        //selectableCards := make([]uint, 0, len(g.Players[user].Hand.Cards)-1)
        //for _, thisCard := range g.Players[user].Hand.Cards {
          //if thisCard.GameID != card.GameID {
            //selectableCards = append(selectableCards, thisCard.GameID)
          //}
        //}
        //empty := (make([]CardMovement, 0))
        //return &UpdateInfo{
          //Movements: movements,
          //Phase: PHASE_SELECTING_CARDS,
          //Pile: HAND_PILE,
          //OpenViewCards: make([]uint, 0),
          //SelectableCards: selectableCards,
        //}, &UpdateInfo{
          //Movements: *mergeMoves(&empty, &movements),
          //Phase: PHASE_OPPONENTS_TURN,
          //Pile: HAND_PILE,
          //OpenViewCards: make([]uint, 0),
          //SelectableCards: make([]uint, 0),
        //}, nil
      } else {
        selectableCards := make([]uint, 0, len(g.Players[user].Hand.Cards))
        for _, thisCard := range g.Players[user].Hand.Cards {
          selectableCards = append(selectableCards, thisCard.GameID)
        }
        movements := append(make([]CardMovement, 0, 1), CardMovement{
          From: HAND_PILE,
          To: DISCARD_PILE,
          GameID: action.SelectedCards[0],
          CardID: card.ID,
        })
        empty := (make([]CardMovement, 0))
        return &UpdateInfo{
          Movements: movements,
          Phase: PHASE_MY_TURN,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: selectableCards,
        }, 
        &UpdateInfo{
          Movements: *mergeMoves(&empty, &movements),
          Phase: PHASE_OPPONENTS_TURN,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: make([]uint, 0),
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
    empty := (make([]CardMovement, 0))
    return &UpdateInfo{
      Movements: movements,
      Phase: PHASE_MY_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: selectableCards,
    }, &UpdateInfo{
      Movements: *mergeMoves(&empty, &movements),
      Phase: PHASE_OPPONENTS_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: make([]uint, 0),
    }, nil
  }

  return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Not sure how to handle action")
}

func MakeGame(cardHandler *CardHandler) *Game {
	return &Game{
		CardIndex: 0,
		Players: make([]Player, 0, 2),
    CardHandler: cardHandler,
    CardActionStack: nil,
	}
}



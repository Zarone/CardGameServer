package gamemanager

import (
	"errors"
	"fmt"
)

type StaticPileData struct {
  publicKnowledge bool
}

type Game struct {
	Players         []Player
	CardIndex       uint
  CardHandler     *CardHandler
  CardActionStack *CardActionStack
  PerPlayerPiles  map[Pile]*StaticPileData
}

func MakeGame(cardHandler *CardHandler) *Game {
	return &Game{
		CardIndex: 0,
		Players: make([]Player, 0, 2),
    CardHandler: cardHandler,
    CardActionStack: nil,
    PerPlayerPiles: map[Pile]*StaticPileData{
      HAND_PILE: {publicKnowledge: false}, 
      DECK_PILE: {publicKnowledge: false}, 
      DISCARD_PILE: {publicKnowledge: true}, 
    },
	}
}

// returns the index of this player within
// the players of this game
func (g *Game) AddPlayer() uint8 {
	var newIndex uint8 = uint8(len(g.Players))
	g.Players = append(g.Players, MakePlayer(g.PerPlayerPiles))
	return newIndex
}

// Sets up player with the given playerID with the deck given by 
// an array of the card IDs.
func (g *Game) SetupPlayer(playerID uint8, deck []uint) {
	var player *Player = &g.Players[playerID]

  playerDeck, ok := g.Players[playerID].PlayerPiles[DECK_PILE]
  if !ok { fmt.Println("Could not find deck pile"); return }

	playerDeck.Cards = make([]Card, 0, len(deck))
	for _, el := range deck {
		playerDeck.Cards = append(playerDeck.Cards, Card{
			ID: el,
			GameID: g.CardIndex,
		})
    player.FindID[g.CardIndex] = playerDeck
		g.CardIndex++
	}
}

// Takes cardIDs, and returns the corresponding game IDs
func (g *Game) GetSetupData(playerID uint8) (*[]uint, *[]uint) {
  playerDeck, ok := g.Players[playerID].PlayerPiles[DECK_PILE]
  if !ok { fmt.Println("Could not find deck pile"); return nil, nil }

  myDeck := make([]uint, 0, len(playerDeck.Cards))
	for _, el := range playerDeck.Cards {
    myDeck = append(myDeck, el.GameID)
	}

  oppPlayerDeck, ok := g.Players[1-playerID].PlayerPiles[DECK_PILE]
  if !ok { fmt.Println("Could not find opponent deck pile"); return nil, nil }

  oppDeck := make([]uint, 0, len(oppPlayerDeck.Cards))
	for _, el := range oppPlayerDeck.Cards {
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
  p1Deck, ok := g.Players[0].PlayerPiles[DECK_PILE]
  if !ok { fmt.Println("Could not find deck pile"); return nil, nil }
  p2Deck, ok := g.Players[1].PlayerPiles[DECK_PILE]
  if !ok { fmt.Println("Could not find deck pile"); return nil, nil }
  p1Hand, ok := g.Players[0].PlayerPiles[HAND_PILE]
  if !ok { fmt.Println("Could not find hand pile"); return nil, nil }
  p2Hand, ok := g.Players[1].PlayerPiles[HAND_PILE]
  if !ok { fmt.Println("Could not find hand pile"); return nil, nil }

	p1Deck.shuffle()
	p2Deck.shuffle()
  fmt.Println(g.Players[0], g.Players[1])
  p1Moves, p2Moves := g.Players[0].moveFromTopTo(p1Deck, p1Hand, 7), 
		g.Players[1].moveFromTopTo(p2Deck, p2Hand, 7)
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
    Movements: *g.mergeMoves(p1Moves, p2Moves),
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
    Movements: *g.mergeMoves(p2Moves, p1Moves),
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
      playerHand, ok := g.Players[user].PlayerPiles[HAND_PILE]
      if !ok { return nil, nil, errors.New("Could not find hand") }
      playerDiscard, ok := g.Players[user].PlayerPiles[DISCARD_PILE]
      if !ok { return nil, nil, errors.New("Could not find discard") }

      card := playerHand.find(action.SelectedCards[0])
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
        return info, g.toOppInfo(info), nil
      } else {
        movements := append(
          make([]CardMovement, 0, 1), 
          g.Players[user].moveCardTo(
            action.SelectedCards[0], 
            playerDiscard,
          ),
        )
        info := &UpdateInfo{
          Movements: movements,
          Phase: PHASE_MY_TURN,
          Pile: HAND_PILE,
          OpenViewCards: make([]uint, 0),
          SelectableCards: *g.getPlayableCards(user),
        }
        return info, g.toOppInfo(info), nil
      }
    }
  } else if ActionType(action.ActionType) == ActionTypeFinishSelection {
    if g.CardActionStack != nil {
      info, _, err := g.processCardAction(user, nil, action, nil)
      if err != nil {
        return nil, nil, err
      }
      return info, g.toOppInfo(info), nil
    }

    playerHand, ok := g.Players[user].PlayerPiles[HAND_PILE]
    if !ok { return nil, nil, errors.New("Could not find hand") }

    movements := make([]CardMovement, 0, len(action.SelectedCards))
    for _, el := range action.SelectedCards {
      movements = append(movements, CardMovement{
        From: HAND_PILE,
        To: DISCARD_PILE,
        GameID: el,
        CardID: playerHand.find(el).ID, 
      })
    }

    selectableCards := make([]uint, 0, len(playerHand.Cards))
    for _, thisCard := range playerHand.Cards {
      selectableCards = append(selectableCards, thisCard.GameID)
    }

    info := &UpdateInfo{
      Movements: movements,
      Phase: PHASE_MY_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: selectableCards,
    }
    return info, g.toOppInfo(info), nil
  }

  return &UpdateInfo{}, &UpdateInfo{}, fmt.Errorf("Not sure how to handle action")
}

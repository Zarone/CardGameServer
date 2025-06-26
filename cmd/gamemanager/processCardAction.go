package gamemanager

import (
	"errors"
	"fmt"
)

func (g *Game) getApplicableCards(user uint8, filter *CardFilter) *[]uint {
  switch filter.Kind {
  case "AND": 
    fmt.Println("UNHANDLED FILTER KIND:", filter.Kind)
    return nil
  case "OR": 
    fmt.Println("UNHANDLED FILTER KIND:", filter.Kind)
    return nil
  case "JUST": 
    cards := make([]uint, 0)

    playerPile, ok := g.Players[user].PlayerPiles[Pile(filter.Pile)]
    if !ok { fmt.Println("Could not find pile", filter.Pile); return nil }

    for _, card := range playerPile.Cards {
      if filter.Type == "" || g.CardHandler.cardLookup["set1"][card.ID].CardType == filter.Type {
        cards = append(cards, card.GameID)
      }
    }
    return &cards
  default: 
    fmt.Println("UNKNOWN FILTER KIND:", filter.Kind)
    return nil
  }
}

func (g *Game) getPlayableCards(user uint8) *[]uint {
  playable := make([]uint, 0)
  playerHand, ok := g.Players[user].PlayerPiles[HAND_PILE]
  if !ok { fmt.Println("Could not find hand"); return nil }
  for _, card := range playerHand.Cards {
    cond := g.CardHandler.cardLookup["set1"][card.ID].PreCondition
    condEval := true
    if cond != nil {
      var err error
      condEval, err = g.evaluateBoolExpression(user, cond)
      if err != nil {
        fmt.Printf("Error evaluating precondition on card with CardID: %d\n", card.ID)
      }
    }

    if condEval {
      playable = append(playable, card.GameID)
    }
  }
  return &playable
}

func (g *Game) processCardAction(user uint8, cardEffect *CardEffect, action *Action, targetToPopulate *[]uint) (*UpdateInfo, bool, error) {
  fromStack := g.CardActionStack != nil
  effect := cardEffect
  incitingAction := action
  startIndex := 0
  if g.CardActionStack != nil {
    startIndex = g.CardActionStack.lastArgument
    effect = g.CardActionStack.lastEffect
    incitingAction = g.CardActionStack.incitingAction
    g.CardActionStack = g.CardActionStack.inner
  }

  fmt.Println("KIND:", effect.Kind)
  switch effect.Kind {
  case "THEN":
    var info UpdateInfo
    info.Movements = make([]CardMovement, 0)
    for i := startIndex; i < len(effect.Args); i++ {
      el := effect.Args[i]
      localInfo, controlReturned, err := g.processCardAction(user, el, action, nil)
      if err != nil {
        return nil, false, err
      }

      info.Pile = localInfo.Pile
      info.SelectableCards = localInfo.SelectableCards
      info.SelectionRestrictions = localInfo.SelectionRestrictions
      info.OpenViewCards = localInfo.OpenViewCards
      info.Phase = localInfo.Phase
      for _, movement := range localInfo.Movements {
        info.Movements = append(info.Movements, movement)
      }

      if controlReturned {
        g.CardActionStack = &CardActionStack{
          lastArgument: i,
          lastEffect: effect,
          inner: g.CardActionStack,
          incitingAction: incitingAction,
        }
        return &info, true, nil
      }
    }
    return &info, false, nil
  case "OR":
    return nil, false, fmt.Errorf("Unhandled Effect Kind: %s\n", effect.Kind)
  case "MOVE":
    var selectedCards []uint
    info, controlReturned, err := g.processCardAction(user, effect.CardTarget, action, &selectedCards)
    if err != nil {
      return &UpdateInfo{}, false, err 
    }
    if controlReturned {
      g.CardActionStack = &CardActionStack{
        lastEffect: effect,
        inner: g.CardActionStack,
        incitingAction: incitingAction,
      }
      return info, true, nil
    }

    fmt.Println("TODO: Validate selected cards fit filter")

    movements := make([]CardMovement, 0)
    for _, cardGameID := range selectedCards {
      group, ok := g.Players[user].PlayerPiles[Pile(effect.To)]
      if !ok { return nil, false, fmt.Errorf("could not find pile %s\n", effect.To) }

      movements = append(
        movements, 
        g.Players[user].moveCardTo(
          cardGameID, 
          group,
        ),
      )
    }

    returnInfo := &UpdateInfo{
      Movements: movements,
      Phase: PHASE_MY_TURN,
      Pile: HAND_PILE,
      OpenViewCards: make([]uint, 0),
      SelectableCards: *g.getPlayableCards(user),
    }
    return returnInfo, false, nil
  case "SHUFFLE":
    return nil, false, fmt.Errorf("Unhandled Effect Kind: %s\n", effect.Kind)
  case "TARGET":
    if fromStack { 
      if targetToPopulate == nil {
        return &UpdateInfo{}, false, errors.New("tried to populate target, but pointer was nil")
      }
      *targetToPopulate = action.SelectedCards

      return &UpdateInfo{}, false, nil 
    }

    if effect.TargetType == "SELECT" {
      g.CardActionStack = &CardActionStack{
        lastEffect: effect,
        inner: g.CardActionStack,
      }

      fmt.Println("Target", effect.Filter.Count)

      return &UpdateInfo{
        Movements: make([]CardMovement, 0),
        Phase: PHASE_SELECTING_CARDS,
        Pile: HAND_PILE,
        OpenViewCards: make([]uint, 0),
        SelectableCards: *g.getApplicableCards(user, &effect.Filter),
        SelectionRestrictions: effect.Filter.Count,
      }, true, nil 
    } else if effect.TargetType == "THIS" {
      if len(incitingAction.SelectedCards) != 1 {
        return &UpdateInfo{}, false, fmt.Errorf("TargetType this, with %d selected cards\n", len(incitingAction.SelectedCards))
      }

      *targetToPopulate = incitingAction.SelectedCards

      return &UpdateInfo{
        Movements: make([]CardMovement, 0),
        Phase: PHASE_MY_TURN,
        Pile: HAND_PILE,
        OpenViewCards: make([]uint, 0),
        SelectableCards: make([]uint, 0),
      }, false, nil 
    } else {
      return nil, false, fmt.Errorf("Unhandled Target Type: %s\n", cardEffect.TargetType)
    }
  default:
    return nil, false, fmt.Errorf("Unknown Effect Kind: %s\n", cardEffect.Kind)
  }

}

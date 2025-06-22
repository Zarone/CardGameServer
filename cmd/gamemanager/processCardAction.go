package gamemanager

import (
	"errors"
	"fmt"
)

func (g *Game) processCardAction(user uint8, cardEffect *CardEffect, action *Action, targetToPopulate *[]uint) (*UpdateInfo, bool, error) {
  fromStack := g.CardActionStack != nil
  effect := cardEffect
  startIndex := 0
  if g.CardActionStack != nil {
    startIndex = g.CardActionStack.lastArgument
    effect = g.CardActionStack.lastEffect
    g.CardActionStack = g.CardActionStack.inner
  }

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
      return info, true, nil
    }

    fmt.Println("TODO: Validate selected cards fit filter")

    movements := make([]CardMovement, 0)
    for _, cardGameID := range selectedCards {
      movements = append(movements, CardMovement{
        GameID: cardGameID,
        From: g.Players[user].FindID[cardGameID].Pile,
        To: StringToPile(effect.To),
      })
    }

    return &UpdateInfo{
      Movements: movements,
    }, false, nil

  case "SHUFFLE":
    return nil, false, fmt.Errorf("Unhandled Effect Kind: %s\n", effect.Kind)
  case "TARGET":
    if fromStack { 
      if targetToPopulate == nil {
        return &UpdateInfo{}, false, errors.New("tried to populate target, but pointer was nil")
      }
      g.CardActionStack = &CardActionStack{
        lastEffect: effect,
        inner: g.CardActionStack,
      }
      *targetToPopulate = action.SelectedCards
      return &UpdateInfo{}, false, nil 
    }

    if effect.TargetType == "SELECT" {
      fmt.Println("TODO: Find selectable cards in select")
      return &UpdateInfo{
        Movements: make([]CardMovement, 0),
        Phase: PHASE_SELECTING_CARDS,
        Pile: HAND_PILE,
        OpenViewCards: make([]uint, 0),
        SelectableCards: make([]uint, 0),
      }, true, nil 
    }

    return nil, false, fmt.Errorf("Unhandled Effect Kind: %s\n", cardEffect.Kind)
  default:
    return nil, false, fmt.Errorf("Unknown Effect Kind: %s\n", cardEffect.Kind)
  }

}

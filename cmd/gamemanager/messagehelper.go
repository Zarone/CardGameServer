package gamemanager

import "fmt"

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

// Takes UpdateInfo from user, and returns the equivalent to
// send to the opponent
func toOppInfo(info *UpdateInfo) *UpdateInfo {
  empty := (make([]CardMovement, 0))
  return &UpdateInfo{
    Movements: *mergeMoves(&empty, &info.Movements),
    Phase: PHASE_OPPONENTS_TURN,
    Pile: HAND_PILE,
    OpenViewCards: make([]uint, 0),
    SelectableCards: make([]uint, 0),
  }
}

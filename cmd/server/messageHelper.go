package server

import (
	"fmt"
  "time"
	"github.com/Zarone/CardGameServer/cmd/gamemanager"
)

func timestamp() string {
	return fmt.Sprint(time.Now().Format("20060102150405"))
}

// Takes a pile and converts it to the opponent's equivalent of that pile
func toOpp(pile gamemanager.Pile) gamemanager.Pile {
  if pile == gamemanager.HAND_PILE {
    return gamemanager.OPP_HAND_PILE
  } else if pile == gamemanager.RESERVE_PILE {
    return gamemanager.OPP_RESERVE_PILE
  } else if pile == gamemanager.SPECIAL_PILE {
    return gamemanager.OPP_SPECIALS_PILE
  } else if pile == gamemanager.BATTLEFIELD_PILE {
    return gamemanager.OPP_BATTLEFIELD_PILE
  } else if pile == gamemanager.DISCARD_PILE {
    return gamemanager.OPP_DISCARD_PILE
  } else if pile == gamemanager.DECK_PILE {
    return gamemanager.OPP_DECK_PILE
  } else {
    fmt.Println("Not sure what opponent's version of this pile is", pile)
    return pile
  }
}

// Takes moves of player and opponent and returns a merged cardMovement slice to send to player
func mergeMoves(thisPlayerMoves *[]gamemanager.CardMovement, oppPlayerMoves *[]gamemanager.CardMovement) *[]gamemanager.CardMovement {
  ret := make([]gamemanager.CardMovement, 0, len(*thisPlayerMoves)+len(*oppPlayerMoves))
  for _, movement := range *thisPlayerMoves {
    ret = append(ret, movement)
  }
  for _, movement := range *oppPlayerMoves{
    ret = append(ret, gamemanager.CardMovement{
      From: toOpp(movement.From),
      To: toOpp(movement.To),
      GameID: movement.GameID,
      CardID: movement.CardID,
    })
  }
  return &ret
}

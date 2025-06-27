package gamemanager

import (
	"fmt"
)


type Player struct {
  PlayerPiles map[Pile]*CardGroup
  FindID  map[uint]*CardGroup
}

func MakePlayer(piles map[Pile]*StaticPileData) Player {
  playerPiles := make(map[Pile]*CardGroup, 0)
  for pile, data := range piles {
    playerPiles[pile] = &CardGroup{
      Cards: make([]Card, 0),
      Pile: pile,
      PublicKnowledge: data.publicKnowledge,
    }
  }

  return Player{
    PlayerPiles: playerPiles,
    FindID: make(map[uint]*CardGroup),
	}
}

func (p Player) String() string {
  return fmt.Sprintf(
    "deck: %d\nhand: %d\n", 
    len(p.PlayerPiles["DECK"].Cards), 
    len(p.PlayerPiles["DISCARD"].Cards),
  )
}

func (p *Player) moveCardTo(gameID uint, to *CardGroup) CardMovement {
  from := p.FindID[gameID]

  // find the index of the card
  card, index := from.findCard(gameID)
  if index == -1 {
    fmt.Println("could not find card with gameid:", gameID)
  }

  // remove from current group
  from.Cards = append(from.Cards[:index], from.Cards[index+1:]...)

  // add to new group
  to.Cards = append(to.Cards, card)

  // update FindID
  p.FindID[gameID] = to

  return CardMovement{
    GameID: gameID,
    From: from.Pile,
    To: to.Pile,
    CardID: card.ID,
  }
}

// Moves "numberOfCards" the top (the end) of given card group into "to"
func (p *Player) moveFromTopTo(from *CardGroup, to *CardGroup, numberOfCards uint) *[]CardMovement {
 if uint(len(from.Cards)) < numberOfCards {
    // Handle the case where there are fewer than requested elements
    newMovements := make([]CardMovement, 0, len(from.Cards))

    for i := range from.Cards {
      newMovements = append(newMovements, CardMovement{
        GameID: uint(from.Cards[i].GameID),
        CardID: uint(from.Cards[i].ID),
        From: from.Pile,
        To: to.Pile,
      })
      p.FindID[from.Cards[i].GameID] = to
    }

    to.Cards = append(to.Cards, (from.Cards)...)
    from.Cards = (from.Cards)[:0] // clear src
  
    return &newMovements
  }

  newMovements := make([]CardMovement, 0, numberOfCards)

  for i := range int(numberOfCards) {
    newMovements = append(newMovements, CardMovement{
      GameID: uint(from.Cards[len(from.Cards)-i-1].GameID),
      CardID: uint(from.Cards[len(from.Cards)-i-1].ID),
      From: from.Pile,
      To: to.Pile,
    })
    p.FindID[from.Cards[len(from.Cards)-i-1].GameID] = to
  }

  to.Cards = append(to.Cards, (from.Cards)[len(from.Cards)-int(numberOfCards):]...)
  from.Cards = (from.Cards)[:len(from.Cards)-int(numberOfCards)]

  return &newMovements
}

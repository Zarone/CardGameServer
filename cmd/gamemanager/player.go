package gamemanager

import (
	"fmt"
	"math/rand"
)

type Card struct {
  ID     uint
  GameID uint
}

func (c Card) String() string {
  return fmt.Sprintf("[ID: %d, GameID: %d]\n", c.ID, c.GameID)
}

type CardGroup struct {
  Cards []Card
  Pile  Pile
}

func (cg *CardGroup) shuffle() {
  rand.Shuffle(len(cg.Cards), func(i, j int) {
    (cg.Cards)[i], (cg.Cards)[j] = (cg.Cards)[j], (cg.Cards)[i]
  })
}

func (cg *CardGroup) moveFromTopTo(to *CardGroup, numberOfCards uint) *[]CardMovement {
 if uint(len(cg.Cards)) < numberOfCards {
    // Handle the case where there are fewer than requested elements
    newMovements := make([]CardMovement, 0, len(cg.Cards))

    for i := 0; i < len(cg.Cards); i++ {
      newMovements = append(newMovements, CardMovement{
        CardID: uint(cg.Cards[i].GameID),
        From: cg.Pile,
        To: to.Pile,
      })
    }

    to.Cards = append(to.Cards, (cg.Cards)...)
    cg.Cards = (cg.Cards)[:0] // clear src

  }

  newMovements := make([]CardMovement, 0, numberOfCards)

  for i := 0; i < int(numberOfCards); i++ {
    newMovements = append(newMovements, CardMovement{
      CardID: uint(cg.Cards[len(cg.Cards)-i-1].GameID),
      From: cg.Pile,
      To: to.Pile,
    })
  }

  to.Cards = append(to.Cards, (cg.Cards)[:numberOfCards]...)
  cg.Cards= (cg.Cards)[numberOfCards:]

  return &newMovements

}

type Player struct {
  Deck CardGroup
  Hand CardGroup
}

func (p *Player) String() string {
  return fmt.Sprintf("deck: %s\n", p.Deck.Cards)
}

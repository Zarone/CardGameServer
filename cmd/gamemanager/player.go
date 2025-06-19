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
  return fmt.Sprintf("[ID: %d, GameID: %d]", c.ID, c.GameID)
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

func (cg *CardGroup) find(gameID uint) *Card {
  for _, card := range cg.Cards {
    if card.GameID == gameID {
      return &card
    }
  }
  return nil
}

func (cg *CardGroup) String() string {
  str := ""
  for _, card := range cg.Cards {
    str += card.String() + "\n"
  }
  return str
}

// Moves "numberOfCards" the top (the end) of given card group into "to"
func (cg *CardGroup) moveFromTopTo(to *CardGroup, numberOfCards uint) *[]CardMovement {
 if uint(len(cg.Cards)) < numberOfCards {
    // Handle the case where there are fewer than requested elements
    newMovements := make([]CardMovement, 0, len(cg.Cards))

    for i := range cg.Cards {
      newMovements = append(newMovements, CardMovement{
        GameID: uint(cg.Cards[i].GameID),
        CardID: uint(cg.Cards[i].ID),
        From: cg.Pile,
        To: to.Pile,
      })
    }

    to.Cards = append(to.Cards, (cg.Cards)...)
    cg.Cards = (cg.Cards)[:0] // clear src
  
    return &newMovements
  }

  newMovements := make([]CardMovement, 0, numberOfCards)

  for i := range int(numberOfCards) {
    newMovements = append(newMovements, CardMovement{
      GameID: uint(cg.Cards[len(cg.Cards)-i-1].GameID),
      CardID: uint(cg.Cards[len(cg.Cards)-i-1].ID),
      From: cg.Pile,
      To: to.Pile,
    })
  }

  to.Cards = append(to.Cards, (cg.Cards)[len(cg.Cards)-int(numberOfCards):]...)
  cg.Cards = (cg.Cards)[:len(cg.Cards)-int(numberOfCards)]

  return &newMovements

}

type Player struct {
  Deck CardGroup
  Hand CardGroup
}

func (p Player) String() string {
  return fmt.Sprintf("deck: %s\nhand: %s\n", p.Deck.String(), p.Hand.String())
}

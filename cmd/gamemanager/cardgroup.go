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

func (cg *CardGroup) findCard (gameID uint) (Card, int) {
  for index, card := range cg.Cards {
    if card.GameID == gameID {
      return card, index
    }
  }
  return Card{}, -1
}

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

type Player struct {
  Deck    CardGroup
  Hand    CardGroup
  Discard CardGroup
  FindID  map[uint]*CardGroup
}

func (p Player) String() string {
  return fmt.Sprintf("deck: %s\nhand: %s\n", p.Deck.String(), p.Hand.String())
}

func MakePlayer() Player {
  return Player{
		Deck: CardGroup{
			Cards: make([]Card, 0),
			Pile: DECK_PILE,
		},
		Hand: CardGroup{
			Cards: make([]Card, 0),
			Pile: HAND_PILE,
		},
    Discard: CardGroup{
      Cards: make([]Card, 0),
      Pile: DISCARD_PILE,
    },
	}
}

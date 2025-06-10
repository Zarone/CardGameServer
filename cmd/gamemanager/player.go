package gamemanager

import "fmt"

type Card struct {
  id int
  gameid int
}

func (c Card) String() string {
  return fmt.Sprintf("[ID: %d, GameID: %d]\n", c.id, c.gameid)
}

type Player struct {
  deck []Card
}

func (p *Player) toString() string {
  return fmt.Sprint(p.deck)
}

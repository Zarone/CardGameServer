package gamemanager

import "fmt"

type Action struct {
}

type Game struct {
  cardIndex int
  players []Player
}

func (g *Game) AddPlayer() uint8 {
  var newIndex uint8 = uint8(len(g.players))
  g.players = append(g.players, Player{
    deck: make([]Card, 0),
  })
  return newIndex
}

func (g *Game) SetupPlayer(playerID uint8, deck []int) {
  var player *Player = &g.players[playerID]
  player.deck = make([]Card, len(deck))
  for index, el := range deck {
    player.deck[index] = Card{
      id: el,
      gameid: g.cardIndex,
    }
    g.cardIndex++
  }

  fmt.Println(g.toString())
}

func (g *Game) toString() string {
  str := "[Game:\n"

  for _, el := range g.players {
    str += el.toString() + ",\n"
  }

  return str + "]"
}

func (g *Game) ProcessAction(user uint8, action *Action) {
}

func MakeGame() *Game {
  return &Game{
    cardIndex: 0,
    players: make([]Player, 0, 2),
  }
}



package gamemanager

type Action struct {
}

type Game struct {
	Players    []Player
	CardIndex  uint
}

// returns the index of this player within
// the players of this game
func (g *Game) AddPlayer() uint8 {
	var newIndex uint8 = uint8(len(g.Players))
	g.Players = append(g.Players, Player{
		Deck: CardGroup{
			Cards: make([]Card, 0),
			Pile: DECK_PILE,
		},
		Hand: CardGroup{
			Cards: make([]Card, 0),
			Pile: HAND_PILE,
		},
	})
	return newIndex
}

// Sets up player with the given playerID with the deck given by 
// an array of the card IDs.
func (g *Game) SetupPlayer(playerID uint8, deck []uint) *[]uint {
	var player *Player = &g.Players[playerID]
	player.Deck.Cards = make([]Card, 0, len(deck))
	gameIdList := make([]uint, 0, len(deck))
	for _, el := range deck {
		player.Deck.Cards = append(player.Deck.Cards, Card{
			ID: el,
			GameID: g.CardIndex,
		})
		gameIdList = append(gameIdList, g.CardIndex)
		g.CardIndex++
	}

	return &gameIdList
}

func (g *Game) String() string {
	str := "[Game:\n"

	for _, el := range g.Players {
		str += el.String() + ",\n"
	}

	return str + "]"
}

func (g *Game) StartGame() (*[]CardMovement, *[]CardMovement) {
	g.Players[0].Deck.shuffle()
	g.Players[1].Deck.shuffle()
	return g.Players[0].Deck.moveFromTopTo(&g.Players[0].Hand, 7), 
		g.Players[1].Deck.moveFromTopTo(&g.Players[1].Hand, 7)
}

func (g *Game) ProcessAction(user uint8, action *Action) {
}

func MakeGame() *Game {
	return &Game{
		CardIndex: 0,
		Players: make([]Player, 0, 2),
	}
}



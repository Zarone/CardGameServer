package server

import (
	"fmt"

	"github.com/Zarone/CardGameServer/cmd/gamemanager"
)


type MessageType string
const (
  MessageTypeSetup = MessageType("SETUP_MESSAGE")
  MessageTypeHeadsOrTails = MessageType("HEADS_OR_TAILS")
  MessageTypeCoinChoice = MessageType("COIN_CHOICE")
  MessageTypeFirstOrSecond = MessageType("FIRST_OR_SECOND")
  MessageTypeFirstOrSecondChoice = MessageType("FIRST_OR_SECOND_CHOICE")
  MessageTypeGameplay = MessageType("GAMEPLAY")
)

type Message[T any] struct {
  Content     T             `json:"content"`
	MessageType MessageType   `json:"type"`
	Timestamp   string        `json:"timestamp"`
}

// Message Content Types
type SetupContent struct {
  Deck []uint `json:"deck"`
}
type CoinFlipContent struct {
  IsChoosingFlip bool `json:"isChoosingFlip"`
}
type CoinFlipContentChoice struct {
  Heads bool `json:"heads"`
}
type StartGameContent struct {
  IsChoosingTurnOrder bool `json:"isChoosingTurnOrder"`
}
type StartGameContentChoice struct {
  First bool `json:"first"`
}
type UpdateInfo struct {
  Movements     []gamemanager.CardMovement  `json:"movements"`
  Phase         uint                        `json:"phase"`
  Pile          gamemanager.Pile            `json:"pile"`
  OpenViewCards []uint                      `json:"openViewCards"`
  MyTurn        bool                        `json:"myTurn"`
}
// gamemanager.Action also counts as one of these
//

func (params *Message[T]) String() string {
  contentString := fmt.Sprint(params.Content) 
  return fmt.Sprintf("[Content: %s, Type: %s, Time: %s]\n", contentString, params.MessageType, params.Timestamp)
}

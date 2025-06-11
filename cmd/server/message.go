package server

import "fmt"

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
  Deck []int `json:"deck"`
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
//

type UpdateInfo struct {
  Movements     []CardMovement  `json:"movements"`
  Phase         uint            `json:"phase"`
  Pile          uint            `json:"pile"`
  OpenViewCards []uint          `json:"openViewCards"`
  MyTurn        bool            `json:"myTurn"`
}

func (params *Message[T]) toString() string {
  contentString := fmt.Sprint(params.Content) 
  return fmt.Sprintf("[Content: %s, Type: %s, Time: %s]\n", contentString, params.MessageType, params.Timestamp)
}

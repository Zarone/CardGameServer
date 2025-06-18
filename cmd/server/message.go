package server

import (
	"fmt"
  "github.com/Zarone/CardGameServer/cmd/gamemanager"
)



type Message[T any] struct {
  Content     T                         `json:"content"`
	MessageType gamemanager.MessageType   `json:"type"`
	Timestamp   string                    `json:"timestamp"`
}

// Message Content Types
type SetupContent struct {
  Deck []uint `json:"deck"`
}
type SetupResponse struct {
  MyDeck  []uint `json:"myDeck"`
  OppDeck []uint `json:"oppDeck"`
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
// gamemanager.UpdateInfo also counts as one of these
// gamemanager.Action also counts as one of these
//

func (params *Message[T]) String() string {
  contentString := fmt.Sprint(params.Content) 
  return fmt.Sprintf("[Content: %s, Type: %s, Time: %s]\n", contentString, params.MessageType, params.Timestamp)
}

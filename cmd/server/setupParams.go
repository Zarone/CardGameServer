package server

import "fmt"

type SetupContent struct {
  Deck []int `json:"deck"`
}

type SetupParams struct {
  Content     SetupContent  `json:"content"`
	MessageType string `json:"type"`
	Timestamp   string `json:"timestamp"`
}

func (params *SetupParams) toString() string {
  contentString := fmt.Sprint(params.Content) 
  return fmt.Sprintf("[Content: %s, Type: %s, Time: %s]\n", contentString, params.MessageType, params.Timestamp)
}


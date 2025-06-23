package gamemanager

import "fmt"

type Action struct {
  ActionType    ActionType  `json:"type"`
  SelectedCards []uint      `json:"selectedCards"`
  From          Pile        `json:"from"`
}

func (a *Action) String() string {
  return fmt.Sprintf("{ActionType: %v, SelectedCards: %v, From: %v}\n", a.ActionType, a.SelectedCards, a.From)
}

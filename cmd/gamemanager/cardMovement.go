package gamemanager

type CardMovement struct {
  CardID  uint `json:"cardId"`
  From    Pile `json:"from"`
  To      Pile `json:"to"`
}


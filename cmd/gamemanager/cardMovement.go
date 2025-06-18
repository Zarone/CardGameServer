package gamemanager

type CardMovement struct {
  GameID  uint `json:"gameId"`
  CardID  uint `json:"cardId"`
  From    Pile `json:"from"`
  To      Pile `json:"to"`
}


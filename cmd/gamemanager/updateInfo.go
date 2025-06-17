package gamemanager

type UpdateInfo struct {
  Movements       []CardMovement  `json:"movements"`
  Phase           Phase           `json:"phase"`
  Pile            Pile            `json:"pile"`
  OpenViewCards   []uint          `json:"openViewCards"`
  MyTurn          bool            `json:"myTurn"`
  SelectableCards []uint          `json:"selectableCards"` 
}

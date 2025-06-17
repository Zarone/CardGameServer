package gamemanager

// These variables should correspond exactly with 
// enums in client code

type Pile uint

const (
  TEMPORARY             = Pile(0)
  HAND_PILE             = Pile(1)
  RESERVE_PILE          = Pile(2)
  SPECIAL_PILE          = Pile(3)
  BATTLEFIELD_PILE      = Pile(4)
  DISCARD_PILE          = Pile(5)
  DECK_PILE             = Pile(6)
  OPP_HAND_PILE         = Pile(7)
  OPP_RESERVE_PILE      = Pile(8)
  OPP_SPECIALS_PILE     = Pile(9)
  OPP_BATTLEFIELD_PILE  = Pile(10)
)

type MessageType uint 

const (
  MessageTypeSetup                = MessageType(0)
  MessageTypeHeadsOrTails         = MessageType(1)
  MessageTypeCoinChoice           = MessageType(2)
  MessageTypeFirstOrSecond        = MessageType(3)
  MessageTypeFirstOrSecondChoice  = MessageType(4)
  MessageTypeGameplay             = MessageType(5)
)

type ActionType uint 

const (
  ActionTypeEndTurn              = ActionType(0)
  ActionTypeSelectCard           = ActionType(1)
  ActionTypeFinishSelection      = ActionType(2)
)

type Phase uint

const (
  PHASE_MY_TURN                   = Phase(0)
  PHASE_OPPONENTS_TURN            = Phase(1)
  PHASE_SELECTING_CARDS           = Phase(2)
  PHASE_SELECTING_TEMPORARY_CARDS = Phase(3)
)

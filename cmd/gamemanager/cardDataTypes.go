package gamemanager

type Expression struct {
	Kind  string  `json:"kind"` // "CONSTANT", "VARIABLE", "OPERATOR"

  // if Kind="CONSTANT"
	Val       int `json:"val,omitempty"`       // for "CONSTANT" and "VARIABLE"

  // if Kind="OPERATOR"
  Operator string         `json:"operator,omitempty"`
	Args     []*Expression  `json:"args,omitempty"`

  // if Kind="VARIABLE"
  Variable string `json:"variable,omitempty"`
}

type CardEffect struct {
  Kind  string  `json:"kind"` // THEN, OR, MOVE, SHUFFLE, TARGET

  // if Kind="THEN" or KIND="OR"
  Args  []*CardEffect `json:"args,omitempty"`

  // if Kind="MOVE"
  CardTarget *CardEffect `json:"target,omitempty"`
  To    string  `json:"to,omitempty"`

  // if Kind="TARGET"
  TargetType string `json:"targetType,omitempty"` // SELECT, ALL, THIS
  Filter CardFilter `json:"filter,omitempty"`
}

type CardFilter struct {
  // AND, OR, JUST
  Kind  string `json:"kind"` 

  // Always Optional
  Count CountRestriction  `json:"count,omitempty"`

  // If Kind="AND" or Kind="OR"
  Args []*CardFilter  `json:"args,omitempty"`

  // If Kind="JUST", Optionally Include These
  Pile  string  `json:"pile,omitempty"`
  Type  string  `json:"type,omitempty"`
  Top   int     `json:"top,omitempty"` // if you wanted to filter for the top 7 cards of deck, for example
}

type CountRestriction struct {
  AtLeast int `json:"atLeast,omitempty"`
  AtMost  int `json:"atMost,omitempty"`
}

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
  Kind  string  `json:"kind"` // THEN, OR, DRAW, DISCARD

  // if Kind="THEN" or KIND="OR"
  Args  []*CardEffect `json:"args,omitempty"`

  // if Kind="DRAW" or KIND="DISCARD"
  CardFilter  CardFilter  `json:"cardFilter,omitempty"`
}

type CardFilter struct {
  // AND, OR, PILE, TYPE, ANY_COMBINATION_OF
  Kind  string `json:"kind"` 

  // If Kind="AND" or Kind="OR" or Kind="ANY_COMBINATION_OF"
  Args []*CardFilter  `json:"args,omitempty"`

  // If Kind="PILE"
  Pile  string  `json:"pile,omitempty"`

  // If Kind="TYPE"
  Type  string  `json:"type,omitempty"`

  // If Kind="PILE" or Kind="TYPE" or Kind="ANY_COMBINATION_OF"
  Count CountRestriction  `json:"count,omitempty"`

}

type CountRestriction struct {
  AT_LEAST  int `json:"atLeast,omitempty"`
  AT_MOST   int `json:"atMost,omitempty"`
}

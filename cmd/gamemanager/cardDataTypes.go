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
  TargetType string `json:"targetType,omitempty"` // SELECT, ALL, TOP, THIS
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
}

type CountRestriction struct {
  AT_LEAST  int `json:"atLeast,omitempty"`
  AT_MOST   int `json:"atMost,omitempty"`
}

// example usage of CardEffect & CardTarget & CardFilter: 
/*

Ultra Ball:

THEN([
  MOVE(THIS, DISCARD),
  MOVE(
    SELECT(JUST(HAND,COUNT(2,2))), 
    DISCARD
  ),
  MOVE(
    SELECT(JUST(DECK,COUNT(1,1),TYPE(POKEMON))),
    HAND
  ),
  SHUFFLE
])


Super Rod:

THEN([
  MOVE(
    SELECT(
      OR(
        JUST(TYPE(POKEMON)),
        JUST(TYPE(ENERGY))
      ), 
      COUNT(1,3)
    ), DECK
  ),
  SHUFFLE
])


Professor's Research:

THEN([
  MOVE(ALL(HAND), DISCARD)
  MOVE(TOP(JUST(DECK), COUNT(7,7)), HAND)
])

*/

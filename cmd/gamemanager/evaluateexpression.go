package gamemanager

import "fmt"

func (g *Game) getGameVariable(user uint8, varName string) (*Expression, error) {
  switch varName {
  case "CARDS_IN_HAND":
    return &Expression{
      Kind: "CONSTANT",
      Val: len(g.Players[user].Hand.Cards),
    }, nil
  default:
    return &Expression{}, fmt.Errorf("UNKNOWN GAME VARIABLE: %s\n", varName)
  }
}

func (g *Game) evaluateOperator(user uint8, expression *Expression) (*Expression, error) {
  switch expression.Operator {
  case ">":
    numArgs := len(expression.Args)
    if numArgs != 2 {
      return nil, fmt.Errorf("Expected 2 arguments to \">\", but received %d\n", numArgs) 
    }
    left, err := g.evaluateToConstant(user, expression.Args[0])
    if err != nil { return nil, err }
    right, err := g.evaluateToConstant(user, expression.Args[1])
    if err != nil { return nil, err }

    val := 0
    if left.Val > right.Val { val = 1 }

    return &Expression{
      Kind: "CONSTANT",
      Val: val,
    }, nil
  default:
    return &Expression{}, fmt.Errorf("UNKNOWN EXPRESSION OPERATOR: %s\n", expression.Operator)
  }
}

func (g *Game) evaluateToConstant(user uint8, expression *Expression) (*Expression, error) {
  switch expression.Kind {
  case "CONSTANT":
    return expression, nil
  case "VARIABLE":
    return g.getGameVariable(user, expression.Variable)
  case "OPERATOR":
    return g.evaluateOperator(user, expression)    
  default:
    return nil, fmt.Errorf("UNKNOWN EXPRESSION KIND: %s\n", expression.Kind)    
  }
}

func (g *Game) evaluateBoolExpression(user uint8, expression *Expression) (bool, error) {
  constExpression, err := g.evaluateToConstant(user, expression)
  if err != nil {
    return false, err
  }
  return constExpression.Val != 0, nil
}

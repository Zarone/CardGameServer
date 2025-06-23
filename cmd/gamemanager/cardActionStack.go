package gamemanager

type CardActionStack struct {
  lastArgument int
  lastEffect *CardEffect
  inner *CardActionStack
  incitingAction *Action
}


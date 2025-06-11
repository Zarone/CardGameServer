package server

import (
  "fmt"
  "log"
  "math/rand"
  "errors"
  "time"
  "encoding/json"
  "github.com/Zarone/CardGameServer/cmd/gamemanager"
  "github.com/Zarone/CardGameServer/cmd/helper"
  "github.com/gorilla/websocket"
)

type User struct {
  conn *websocket.Conn
  isSpectator bool
}

type CoinFlip uint8
const (
  CoinFlipUnset = CoinFlip(0)
  CoinFlipHead  = CoinFlip(1)
  CoinFlipTail  = CoinFlip(2)
)

type Room struct {
  connections map[*User]bool
  playerToGamePlayerID map[*User]uint8
  game *gamemanager.Game
  readyPlayers []*User
  awaitingAllReady chan bool
  expectingCoinFlip CoinFlip
}

const playersToStartGame uint8 = 2

// returns the number of players in the room,
// excluding spectators
func (r *Room) getPlayersInRoom() uint8 {
  var num uint8 = 0

  for player, isActive := range r.connections {
    if !player.isSpectator && isActive {
      num++
    }
  }

  return num
}

func (r *Room) initPlayer(user *User) error {
  fmt.Println("Consider race condition for r.readyPlayers")

  if len(r.readyPlayers) >= int(playersToStartGame) {
    return errors.New("Too many players")
  }

  r.playerToGamePlayerID[user] = r.game.AddPlayer() 
  r.readyPlayers = append(r.readyPlayers, user)

  return nil
}

func (r *Room) checkAllReady() {
  if len(r.readyPlayers) == int(playersToStartGame) {
    for i := 0; i < int(playersToStartGame); i++ {
      r.awaitingAllReady <- true
    }
  }
}

func (r *Room) toString() string {
  str := ""
  for user, isPresent := range r.connections {
    str += fmt.Sprintf("[ConnectionPointer: %p, isPresent: %t, isSpectator: %t], ", user.conn, isPresent, user.isSpectator)
  }
  return str
}

// Sends the deck list to the game state manager to 
// set it up
func (r *Room) sendSetupData(u *User, deck []int) {
  r.game.SetupPlayer(r.playerToGamePlayerID[u], deck)
}

// Takes info from client regarding their deck list 
// and such
func (r *Room) readSetupParams(user *User) (*Message[SetupContent], error) {
  if (user.isSpectator) { return nil, nil }

  defer r.checkAllReady()

  // read in a message
  _, p, err := user.conn.ReadMessage()
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Error Reading Message {%s}", err))
  }

  // print out that message for clarity
  log.Printf("Setup Message: %s", string(p))

  var params Message[SetupContent] 
  
  if err := json.Unmarshal(p, &params); err != nil {
    return nil, errors.New(fmt.Sprintf("Error parsing JSON: %s", err))
  }

  if (params.MessageType != MessageTypeSetup) {
    return nil, errors.New(fmt.Sprintf("Error setting up, message type is %s", params.MessageType))
  }

  log.Println("Setup Message (Parsed):", params.toString())

  if err := user.conn.WriteMessage(websocket.TextMessage, p); err != nil {
    return nil, errors.New(fmt.Sprintf("Error writing message: %s", err))
  }

  return &params, nil
}

// Attempts to remove connection to the room specified by the request
func (r *Room) removeFromRoom(user *User) {
  if len(r.connections) == 0 {
    log.Println("Error removing from room")
    return
  }

  if (!user.isSpectator) {
    fmt.Println("Implement player disconnect")
  }

  r.connections[user] = false

  helper.DebugPrint(r.toString())
}

func (r *Room) readForActions(ws *websocket.Conn) (gamemanager.Action, error) {
  // read in a message
  messageType, p, err := ws.ReadMessage()
  if err != nil {
    return gamemanager.Action{}, errors.New(fmt.Sprintf("Error Reading Message {%s}", err))
  }

  // print out that message for clarity
  log.Printf("Message: %s", string(p))

  if err := ws.WriteMessage(messageType, p); err != nil {
    return gamemanager.Action{}, errors.New(fmt.Sprintf("Error writing message %s", err))
  }

  return gamemanager.Action{}, nil
}

func (r *Room) processAction(user *User, action *gamemanager.Action) {
  r.game.ProcessAction(r.playerToGamePlayerID[user], action)
}

func (r *Room) spectatorLoop(user *User) {
  for {
    _, err := r.readForActions(user.conn)

    if (err != nil) {
      log.Println("Stopped reading from user, endcode: ", err)
      break
    }
  }
}

type CardMovement struct {
  CardID  uint `json:"cardId"`
  From    uint `json:"from"`
  To      uint `json:"to"`
}

// returns (true, nil) if player 1 is going first
func (r *Room) askTurnOrder() (bool, error) {
  isHeads := rand.Intn(2) == 1 

  var userChoosingFlip *User
  var userWaiting *User
  if (isHeads == (r.expectingCoinFlip == CoinFlipHead)) {
    // if player 1 was right, then player 1 chooses
    userChoosingFlip = r.readyPlayers[0]
    userWaiting = r.readyPlayers[1]
  } else {
    // if player 1 was wrong, then player 2 chooses
    userChoosingFlip = r.readyPlayers[1]
    userWaiting = r.readyPlayers[0]
  }

  userChoosingFlip.conn.WriteJSON(Message[StartGameContent]{
    Content: StartGameContent {
      IsChoosingTurnOrder: true,
    },
    MessageType: MessageTypeHeadsOrTails,
  })
  userWaiting.conn.WriteJSON(Message[StartGameContent]{
    Content: StartGameContent {
      IsChoosingTurnOrder: false,
    },
    MessageType: MessageTypeFirstOrSecond,
  })

  var decisionResponse Message[StartGameContentChoice]

  userChoosingFlip.conn.ReadJSON(decisionResponse)
  if decisionResponse.MessageType != MessageTypeFirstOrSecondChoice {
    return false, errors.New(
      "Client response was expected to be a first or second choice, but was instead " + 
      string(decisionResponse.MessageType),
    )
  }
  
  return (userChoosingFlip == r.readyPlayers[0]) == decisionResponse.Content.First, nil
}

func (r *Room) startGame(user *User) error {
  if (r.playerToGamePlayerID[user] == 0) {
    // this player chooses heads or tails
    var update Message[CoinFlipContent] = Message[CoinFlipContent]{
      Content: CoinFlipContent {
        IsChoosingFlip: true,
      },
      MessageType: MessageTypeHeadsOrTails,
      Timestamp: fmt.Sprint(time.Now().Format("20060102150405")),
    }
    user.conn.WriteJSON(update)
    var decisionResponse Message[CoinFlipContentChoice]

    user.conn.ReadJSON(decisionResponse)
    if decisionResponse.MessageType != MessageTypeCoinChoice {
      return errors.New(
        "Client response was expected to be a coin choice, but was instead " + 
        string(decisionResponse.MessageType),
      )
    }

    if decisionResponse.Content.Heads {
      r.expectingCoinFlip = CoinFlipHead
    } else {
      r.expectingCoinFlip = CoinFlipTail
    }

    for i := 0; i < int(playersToStartGame); i++ {
      r.awaitingAllReady <- true
    }

  } else if (r.playerToGamePlayerID[user] == 1) {
    // this player waits
    var update Message[CoinFlipContent] = Message[CoinFlipContent]{
      Content: CoinFlipContent {
        IsChoosingFlip: false,
      },
      MessageType: MessageTypeHeadsOrTails,
      Timestamp: fmt.Sprint(time.Now().Format("20060102150405")),
    }
    user.conn.WriteJSON(update)
  } else {
    log.Println("Haven't handled scenario with more than 2 players")
  }

  // wait until coin flip decided
  <- r.awaitingAllReady

  if (r.expectingCoinFlip == CoinFlipUnset) {
    return errors.New("Coin Flip isn't set by evaluation time")
  }
  
  // arbitrarily let player 1 execute the following below,
  // or rather let the server call initiated by player 1 
  // execute the below code
  if r.playerToGamePlayerID[user] != 0 { 
    <- r.awaitingAllReady
    return nil
  }

  goingFirst, err := r.askTurnOrder()
  if err != nil {
    return err
  }

  r.readyPlayers[0].conn.WriteJSON(Message[UpdateInfo]{
    Content: UpdateInfo{
      Movements: make([]CardMovement, 0),
      Phase: 0,
      Pile: 0,
      OpenViewCards: make([]uint, 0),
      MyTurn: goingFirst,
    },
    MessageType: MessageTypeGameplay,
    Timestamp: fmt.Sprint(time.Now().Format("20060102150405")),
  })
  r.readyPlayers[1].conn.WriteJSON(Message[UpdateInfo]{
    Content: UpdateInfo{
      Movements: make([]CardMovement, 0),
      Phase: 0,
      Pile: 0,
      OpenViewCards: make([]uint, 0),
      MyTurn: !goingFirst,
    },
    MessageType: MessageTypeGameplay,
    Timestamp: fmt.Sprint(time.Now().Format("20060102150405")),
  })

  for i := 0; i < int(playersToStartGame)-1; i++ {
    r.awaitingAllReady <- true
  }

  return nil
}

func (r *Room) playerLoop(user *User) {
  for {
    action, err := r.readForActions(user.conn)

    if (err != nil) {
      log.Println("Stopped reading from user, endcode: ", err)
      break
    }

    r.processAction(user, &action)
  }
}

func makeRoom() *Room {
  return &Room{
    playerToGamePlayerID: make(map[*User]uint8),
    connections: make(map[*User]bool),
    game: gamemanager.MakeGame(),
    readyPlayers: make([]*User, 0),
    awaitingAllReady: make(chan bool, playersToStartGame),
    expectingCoinFlip: CoinFlipUnset,
  }
}

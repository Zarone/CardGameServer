package server

import (
  "fmt"
  "log"
  "errors"
  "encoding/json"
  "github.com/Zarone/CardGameServer/cmd/gamemanager"
  "github.com/Zarone/CardGameServer/cmd/helper"
  "github.com/gorilla/websocket"
)

type User struct {
  conn *websocket.Conn
  isSpectator bool
}

type Room struct {
  connections map[*User]bool
  playerToGamePlayerID map[*User]uint8
  game *gamemanager.Game
  readyPlayers uint8
  awaitingAllReady chan bool
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

  if r.readyPlayers >= playersToStartGame {
    return errors.New("Too many players")
  }

  r.playerToGamePlayerID[user] = r.game.AddPlayer() 
  r.readyPlayers++

  return nil
}

func (r *Room) checkAllReady() {
  if r.readyPlayers == playersToStartGame {
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
func (r *Room) readSetupParams(user *User) (*SetupParams, error) {
  if (user.isSpectator) { return nil, nil }

  defer r.checkAllReady()

  // read in a message
  _, p, err := user.conn.ReadMessage()
  if err != nil {
    return nil, errors.New(fmt.Sprintf("Error Reading Message {%s}", err))
  }

  // print out that message for clarity
  log.Printf("Setup Message: %s", string(p))

  var params SetupParams
  
  if err := json.Unmarshal(p, &params); err != nil {
    return nil, errors.New(fmt.Sprintf("Error parsing JSON: %s", err))
  }

  if (params.MessageType != "setup message") {
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

func (r *Room) playerLoop(user *User, params *SetupParams) {
  r.sendSetupData(user, params.Content.Deck)

  <- r.awaitingAllReady

  fmt.Println("Starting game for player")

  if (r.playerToGamePlayerID[user] == 0) {
    // this player chooses heads or tails
  } else if (r.playerToGamePlayerID[user] == 1) {
    // this player waits
  } else {
    log.Println("Haven't handled scenario with more than 2 players")
  }

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
    readyPlayers: 0,
    awaitingAllReady: make(chan bool, playersToStartGame),
  }
}

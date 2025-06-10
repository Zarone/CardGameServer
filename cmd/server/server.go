package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Zarone/CardGameServer/cmd/helper"
)

type Server struct {
  rooms map[uint8]*Room
  settings ServerSettings
}

func (s *Server) toString() string {
  str := "[Server: \n" +
    "    " + s.settings.toString()
  
  for i, room := range s.rooms {
    str += fmt.Sprintf("\n    Room %d: %s", i, room.toString())
  }

  str += "\n]"

  return str
}

// Attempts to add connection to the room specified by the request
func (s *Server) addToRoom(req *http.Request, user *User) (*Room, error) {
  roomNum := requestToRoomNumber(req)

  if (s.rooms[roomNum] == nil) { s.rooms[roomNum] = makeRoom() }

  thisRoom := s.rooms[roomNum]

  if thisRoom.getPlayersInRoom() >= playersToStartGame {
    errorString := fmt.Sprintf("Can't join. Too many players in room %d\n", roomNum)
    return thisRoom, errors.New(errorString)
  } else if !user.isSpectator {
    err := thisRoom.initPlayer(user)
    if (err != nil) { 
      log.Println(err)
      return thisRoom, nil
    }
  }

  s.rooms[roomNum].connections[user] = true

  helper.DebugPrint(s.toString())

  return thisRoom, nil
}

func (s *Server) HandleWS(res http.ResponseWriter, req *http.Request) {
  ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
    log.Printf("Error upgrading: %s", err)
    return
	}

  user := User {
    conn: ws,
    isSpectator: req.URL.Query().Get("spectator") == "true",
  }

  room, err := s.addToRoom(req, &user)
  defer room.removeFromRoom(&user)

  if err != nil {
    log.Printf("Error adding to room: %s", err)
    return
  }

  log.Printf("Client [%p] Connected\n", ws)
  err = ws.WriteMessage(1, []byte("Hi Client!"))
  if err != nil {
    log.Printf("Error writing on client connection: %s", err)
    return
  }

  params, err := room.readSetupParams(&user)

  if err != nil {
    log.Printf("Error reading setup parameters %s", err)
    return
  }

  if user.isSpectator {
    room.spectatorLoop(&user)
  } else {
    room.playerLoop(&user, params)
  }

}

func MakeServer(settings *ServerSettings) *Server {
  return &Server{
    rooms: make(map[uint8]*Room),
    settings: *settings,
  }
}


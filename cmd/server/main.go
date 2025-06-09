package main

import (
  "fmt"
  "log"
  "strconv"
  "net/http"
  "github.com/gorilla/websocket"
)

const DEBUG = true
func debugPrint(data any) {
  fmt.Println(data)
}

type User struct {
  conn *websocket.Conn
}

type Room struct {
  connections map[*User]bool
}

type ServerSettings struct {
}

func (settings *ServerSettings) toString() string {
  return "[ServerSettings: ]"
}

type Server struct {
  rooms map[uint8]Room
  settings ServerSettings
}

func (s *Server) toString() string {
  str := "[Server: \n" +
    s.settings.toString()
  
  for i, room := range s.rooms {
    str += fmt.Sprintf("\n Room %d: ", i)
    for user, isPresent := range room.connections {
      if isPresent { 
        str += "[" + fmt.Sprintf("ConnectionPointer: %p", user.conn) + "], "
      }
    }
  }

  str += "\n]"

  return str
}

// Attempts to add connection to the room specified by the request
func (s *Server) addToRoom(req *http.Request, user *User) {
  if len(s.rooms) >= 2 {

  }

  roomString := req.URL.Query().Get("room")
  var roomNum uint8
  if (roomString == "") {
    roomNum = 255
    fmt.Printf("No room selected, joining default room (%d)\n", roomNum)
  } else {
    parsedRoomString, err := strconv.ParseInt(roomString, 10, 8)
    if err != nil {
      fmt.Printf("Error reading room number %s", roomString);
    }
    roomNum = uint8(parsedRoomString)
    fmt.Printf("Requested room %s\n", roomString)
  }

  if len(s.rooms) == 0 {
    s.rooms[roomNum] = Room{
      connections: make(map[*User]bool),
    }
  }

  s.rooms[roomNum].connections[user] = true

  debugPrint(s.toString())
}

// Attempts to remove connection to the room specified by the request
func (s *Server) removeFromRoom(req *http.Request, user *User) {
  roomString := req.URL.Query().Get("room")
  var roomNum uint8
  if (roomString == "") {
    roomNum = 255
    fmt.Printf("No room selected, joining default room (%d)\n", roomNum)
  } else {
    parsedRoomString, err := strconv.ParseInt(roomString, 10, 8)
    if err != nil {
      fmt.Printf("Error reading room number %s", roomString);
    }
    roomNum = uint8(parsedRoomString)
    fmt.Printf("Requested room %s\n", roomString)
  }

  if len(s.rooms) == 0 {
    log.Printf("Error removing from room %s\n", roomString)
    return
  }

  s.rooms[roomNum].connections[user] = false

  debugPrint(s.toString())
}


func (s *Server) readLoop(ws *websocket.Conn){
  for {
		// read in a message
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error Reading Message {%s}", err)
			return
		}

		// print out that message for clarity
    log.Printf("Message: %s", string(p))

		if err := ws.WriteMessage(messageType, p); err != nil {
			log.Printf("Error writing message %s", err)
			return
		}
	}
}

func (s *Server) handleWS(res http.ResponseWriter, req *http.Request) {
  ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
    log.Printf("Error upgrading: %s", err)
    return
	}

  user := User {
    conn: ws,
  }

  s.addToRoom(req, &user)

	log.Printf("Client [%p] Connected\n", ws)
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
    log.Printf("Error writing on client connection: %s", err)
    return
	}

  s.readLoop(ws)

  s.removeFromRoom(req, &user)
}

func makeServer(settings *ServerSettings) *Server {
  return &Server{
    rooms: make(map[uint8]Room),
    settings: *settings,
  }
}

func main() {
  server := makeServer(&ServerSettings{}) 

  // example path: /socket?room=3&spectator=0
  http.HandleFunc("/socket", server.handleWS)

  fmt.Print("Hello from Server\n")
  log.Fatal(http.ListenAndServe(":3000", nil))
}

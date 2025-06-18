package server

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Server struct {
	Rooms map[uint8]*Room
	settings ServerSettings
}

func (s *Server) String() string {
	str := "[Server: \n" +
		"    " + s.settings.toString()
	
	for i, room := range s.Rooms {
		str += fmt.Sprintf("\n    Room %d: %s", i, room)
	}

	str += "\n]"

	return str
}

// AddToRoom attempts to add connection to the room specified by the request
func (s *Server) AddToRoom(req *http.Request, user *User) (*Room, error) {
	roomNum := requestToRoomNumber(req)

	if (s.Rooms[roomNum] == nil) { s.Rooms[roomNum] = MakeRoom(roomNum) }

	thisRoom := s.Rooms[roomNum]

	if thisRoom.GetPlayersInRoom() >= PlayersToStartGame {
		errorString := fmt.Sprintf("Can't join. Too many players in room %d\n", roomNum)
		return thisRoom, errors.New(errorString)
	} else if !user.IsSpectator {
		err := thisRoom.InitPlayer(user)
		if (err != nil) { 
			log.Println(err)
			return thisRoom, nil
		}
	}

	s.Rooms[roomNum].Connections[user] = true

	return thisRoom, nil
}

func (s *Server) RemoveUserFromRoom(user *User, room *Room) {
	room.RemoveFromRoom(user)
}

func (s *Server) HandleWS(res http.ResponseWriter, req *http.Request) {
	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Printf("Error upgrading: %s", err)
		return
	}

	user := User{
		Conn: ws,
		IsSpectator: req.URL.Query().Get("spectator") == "true",
	}

	room, err := s.AddToRoom(req, &user)
	if err != nil {
		log.Printf("Error adding to room: %s", err)
		return
	}

	defer s.RemoveUserFromRoom(&user, room)

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

	if user.IsSpectator {
		room.spectatorLoop(&user)
	} else {
    room.initGameData(&user, params.Content.Deck)

		// Wait for all players to finish initialization
		room.wait("Finished initialization...")

		var setupResponseMessage Message[SetupResponse] = room.getInitData(&user)

		if err := user.Conn.WriteJSON(setupResponseMessage); err != nil {
			fmt.Println(errors.New(fmt.Sprintf("Error writing message: %s", err)))
			return 
		}

		err := room.startGame(&user)
		if err != nil {
			log.Printf("Error starting game: %s", err)
			return
		}

		room.playerLoop(&user)
	}
}

func (s *Server) HandleRoomsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join("cmd", "server", "templates", "rooms.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (s *Server) HandleRoomsAPI(w http.ResponseWriter, r *http.Request) {
	// Generate HTML for the rooms
	var html string
	for _, room := range s.Rooms {
		html += fmt.Sprintf(`
			<div class="room">
				<h2>Room %d (%s)</h2>
				<ul class="user-list">`, 
			room.RoomNumber, 
			room.RoomDescription,
		)
		
		for player, isActive := range room.Connections {
			status := "Player"

			if player.IsSpectator {
				status = "Spectator"
			}

			if isActive {
				status += " (Active)"
			} else {
				status += " (Not Active)"
			}

			spectatorText := ""
			
			if player.IsSpectator { spectatorText = "spectator" }

			html += fmt.Sprintf(`
				<li class="%s">%s</li>`, 
				spectatorText,
				status,
			)
		}
		
		html += `
				</ul>
			</div>`
	}
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func MakeServer(settings *ServerSettings) *Server {
	return &Server{
		Rooms: make(map[uint8]*Room),
		settings: *settings,
	}
}


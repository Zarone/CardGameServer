package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Zarone/CardGameServer/cmd/gamemanager"
	"github.com/Zarone/CardGameServer/cmd/server"
	"github.com/gorilla/websocket"
)

func TestMakeServer(t *testing.T) {
	settings := &server.ServerSettings{}
	s := server.MakeServer(settings)
	
	if s == nil {
		t.Error("MakeServer returned nil")
	}
	
	if s.Rooms == nil {
		t.Error("Server rooms map is nil")
	}
	
	if len(s.Rooms) != 0 {
		t.Error("Server rooms map should be empty on creation")
	}
}

func TestServerAddToRoom(t *testing.T) {
	settings := &server.ServerSettings{}
	s := server.MakeServer(settings)
	
	// Create a test request
	req := httptest.NewRequest("GET", "/ws?room=1", nil)
	user := &server.User{
		IsSpectator: false,
	}
	
	room, err := s.AddToRoom(req, user)
	
	if err != nil {
		t.Errorf("addToRoom failed: %v", err)
	}
	
	if room == nil {
		t.Error("addToRoom returned nil room")
	}
	
	if room.RoomNumber != 1 {
		t.Errorf("Expected room number 1, got %d", room.RoomNumber)
	}
	
	// Test adding to full room
	for i := 0; i < int(server.PlayersToStartGame); i++ {
		user := &server.User{
			IsSpectator: false,
		}
		_, err := s.AddToRoom(req, user)
		if i >= int(server.PlayersToStartGame)-1 && err == nil {
			t.Error("Expected error when room is full, got nil")
		}
	}
}

func TestServerRemoveUserFromRoom(t *testing.T) {
	settings := &server.ServerSettings{}
	s := server.MakeServer(settings)
	
	// Create a test request
	req := httptest.NewRequest("GET", "/ws?room=1", nil)
	user := &server.User{
		IsSpectator: false,
	}
	
	room, _ := s.AddToRoom(req, user)
	
	// Test removal
	s.RemoveUserFromRoom(user, room)
	
	if room.Connections[user] {
		t.Error("User should be marked as inactive after removal")
	}
}
func TestServerHandleRoomsAPI(t *testing.T) {
	settings := &server.ServerSettings{}
	s := server.MakeServer(settings)
	
	// Add some test rooms
	req1 := httptest.NewRequest("GET", "/ws?room=1", nil)
	user1 := &server.User{IsSpectator: false}
	s.AddToRoom(req1, user1)
	
	req2 := httptest.NewRequest("GET", "/ws?room=2", nil)
	user2 := &server.User{IsSpectator: true}
	s.AddToRoom(req2, user2)
	
	// Test the API endpoint
	w := httptest.NewRecorder()
	s.HandleRoomsAPI(w, httptest.NewRequest("GET", "/api/rooms", nil))
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	
	// Check if the response contains room information
	body := w.Body.String()
	if body == "" {
		t.Error("Empty response from HandleRoomsAPI")
	}
}

// --- Generalized Race Condition Test Harness ---

// SetupPhaseAction represents an action a player can take during setup
// ActionType: "deck", "coin", "turn"
type SetupPhaseAction struct {
	Player int    // 1 or 2
	Type   string // "deck", "coin", "turn"
}

// SimulateSetupPhase runs a sequence of actions (possibly concurrently) for two clients
func SimulateSetupPhase(t *testing.T, actions []SetupPhaseAction) (*server.Room, *websocket.Conn, *websocket.Conn) {
	settings := &server.ServerSettings{}
	s := server.MakeServer(settings)
	ts := httptest.NewServer(http.HandlerFunc(s.HandleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?room=1"

	// Connect both players
	ws := make([]*websocket.Conn, 2)
	for i := 0; i < 2; i++ {
		var err error
		ws[i], _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("WebSocket dial failed for player %d: %v", i+1, err)
		}
		defer ws[i].Close()
		_, p, err := ws[i].ReadMessage()
		if err != nil || string(p) != "Hi Client!" {
			t.Fatalf("Player %d did not receive correct init message: %v, %q", i+1, err, string(p))
		}
	}

	// Prepare messages
	setupMsg := []server.Message[server.SetupContent]{
		{
			Content:      server.SetupContent{Deck: []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			MessageType:  gamemanager.MessageTypeSetup,
			Timestamp:    "test",
		},
		{
			Content:      server.SetupContent{Deck: []uint{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
			MessageType:  gamemanager.MessageTypeSetup,
			Timestamp:    "test",
		},
	}
	coinMsg := []server.Message[server.CoinFlipContentChoice]{
		{
			Content:      server.CoinFlipContentChoice{Heads: true},
			MessageType:  gamemanager.MessageTypeCoinChoice,
			Timestamp:    "test",
		},
		{
			Timestamp: "never",
		},
	}
	turnMsg := []server.Message[server.StartGameContentChoice]{
		{
			Content:      server.StartGameContentChoice{First: true},
			MessageType:  gamemanager.MessageTypeFirstOrSecondChoice,
			Timestamp:    "test",
		},
		{
			Content:      server.StartGameContentChoice{First: true},
			MessageType:  gamemanager.MessageTypeFirstOrSecondChoice,
			Timestamp:    "test",
		},
	}

	// Channels to collect errors
	errCh := make(chan error, 2) // one per player
	var wg sync.WaitGroup

	// Collect actions for each player in order
	playerActions := [][]SetupPhaseAction{ {}, {} }
	for _, act := range actions {
		idx := act.Player - 1
		playerActions[idx] = append(playerActions[idx], act)
	}

	// Launch a goroutine per player to execute their actions in order
	for idx := range 2 {
		wg.Add(1)
		go func(innerIndex int) {
			defer wg.Done()
			for _, act := range playerActions[innerIndex] {
				var err error
				switch act.Type {
				case "deck":
					if setupMsg[innerIndex].Timestamp != "never" {
						fmt.Printf("Player %d: Wrote message: %+v\n", innerIndex+1, setupMsg[innerIndex])
						err = ws[innerIndex].WriteJSON(setupMsg[innerIndex])
					}
				case "coin":
					if coinMsg[innerIndex].Timestamp != "never" {
						fmt.Printf("Player %d: Wrote message: %+v\n", innerIndex+1, coinMsg[innerIndex])
						err = ws[innerIndex].WriteJSON(coinMsg[innerIndex])
					}
				case "turn":
					if turnMsg[innerIndex].Timestamp != "never" {
						fmt.Printf("Player %d: Wrote message: %+v\n", innerIndex+1, turnMsg[innerIndex])
						err = ws[innerIndex].WriteJSON(turnMsg[innerIndex])
					}
				}
				if err != nil {
					errCh <- err
					return
				}
				switch act.Type {
				case "deck":
					var resp server.Message[server.SetupResponse]
					ws[innerIndex].SetReadDeadline(time.Now().Add(time.Second))
					readErr := ws[innerIndex].ReadJSON(&resp)
					if readErr == nil {
						fmt.Printf("Player %d: Received message: %+v\n", innerIndex+1, resp)
					} else {
						fmt.Printf("Player %d: Error reading message: %v\n", innerIndex+1, readErr)
					}
					var respFlipsChoice server.Message[server.CoinFlipContent]
					ws[innerIndex].SetReadDeadline(time.Now().Add(time.Second))
					readErr = ws[innerIndex].ReadJSON(&respFlipsChoice)
					if readErr == nil {
						fmt.Printf("Player %d: Received message: %+v\n", innerIndex+1, respFlipsChoice)
					} else {
						fmt.Printf("Player %d: Error reading message: %v\n", innerIndex+1, readErr)
					}
				case "coin":
					var resp server.Message[server.StartGameContent]
					ws[innerIndex].SetReadDeadline(time.Now().Add(time.Second))
					readErr := ws[innerIndex].ReadJSON(&resp)
					if readErr == nil {
						fmt.Printf("Player %d: Received message: %+v\n", innerIndex+1, resp)
						if act.Type == "coin" && !resp.Content.IsChoosingTurnOrder {
							turnMsg[innerIndex].Timestamp = "never"
						}
					} else {
						fmt.Printf("Player %d: Error reading message: %v\n", innerIndex+1, readErr)
					}
				case "turn":
					var resp server.Message[gamemanager.UpdateInfo]
					ws[innerIndex].SetReadDeadline(time.Now().Add(time.Second))
					readErr := ws[innerIndex].ReadJSON(&resp)
					if readErr == nil {
						fmt.Printf("Player %d: Received message: %+v\n", innerIndex+1, resp)
					} else {
						fmt.Printf("Player %d: Error reading message: %v\n", innerIndex+1, readErr)
					}
				}
			}
		}(idx)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("Setup phase error: %v", err)
		}
	}

	return s.Rooms[1], ws[0], ws[1]
}

// Test permutation (deck1, deck2)
func TestSetupRace_Deck1Deck2(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("TestSetupRace_Deck1Deck2, run-%d", i), func(t *testing.T) {
			r, _, _ := SimulateSetupPhase(t, []SetupPhaseAction{
				{Player: 1, Type: "deck"},
				{Player: 2, Type: "deck"},
			})
			if r.RoomDescription != server.DESC_FINISHED_INITIALIZATION {
				t.Fatalf("Room description was %v", r.RoomDescription)
			}
		})
	}
}

// Test permutation (deck1, deck2, coin1 )
func TestSetupRace_Deck1Deck2Coin1(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("TestSetupRace_Deck1Deck2Coin1, run-%d", i), func(t *testing.T) {
			r, _, _ := SimulateSetupPhase(t, []SetupPhaseAction{
				{Player: 1, Type: "deck"},
				{Player: 2, Type: "deck"},
				{Player: 1, Type: "coin"},
			})
			if r.RoomDescription != server.DESC_HEADS_OR_TAILS_CHOSEN {
				t.Fatalf("Room description was %v", r.RoomDescription)
			}
		})
	}
}

// Test permutation (deck1, deck2, coin1, coin2, turn1, turn2 )
func TestSetupRace_Deck1Deck2Coin1Coin2Turn1Turn2(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("TestSetupRace_Deck1Deck2Coin1Coin2Turn1Turn2, run-%d", i), func(t *testing.T) {
			r, _, _ := SimulateSetupPhase(t, []SetupPhaseAction{
				{Player: 1, Type: "deck"},
				{Player: 2, Type: "deck"},
				{Player: 1, Type: "coin"},
				{Player: 2, Type: "coin"},
				{Player: 1, Type: "turn"},
				{Player: 2, Type: "turn"},
			})
			if r.RoomDescription != server.DESC_INITIAL_STATE_TO_CLIENT {
				t.Fatalf("Room description was %v", r.RoomDescription)
			}
		})
	}
}

package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zarone/CardGameServer/cmd/server"
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
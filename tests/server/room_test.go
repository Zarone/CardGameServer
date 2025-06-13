package server_test

import (
	"testing"

	"github.com/Zarone/CardGameServer/cmd/server"
)

func TestMakeRoom(t *testing.T) {
	roomNum := uint8(1)
	room := server.MakeRoom(roomNum)
	
	if room == nil {
		t.Error("makeRoom returned nil")
	}
	
	if room.RoomNumber != roomNum {
		t.Errorf("Expected room number %d, got %d", roomNum, room.RoomNumber)
	}
	
	if room.Connections == nil {
		t.Error("Room connections map is nil")
	}
	
	if room.PlayerToGamePlayerID == nil {
		t.Error("Room playerToGamePlayerID map is nil")
	}
	
	if room.ReadyPlayers == nil {
		t.Error("Room readyPlayers slice is nil")
	}
	
	if room.AwaitingAllReady == nil {
		t.Error("Room awaitingAllReady channel is nil")
	}
}

func TestRoomGetPlayersInRoom(t *testing.T) {
	room := server.MakeRoom(1)
	
	// Test empty room
	if count := room.GetPlayersInRoom(); count != 0 {
		t.Errorf("Expected 0 players, got %d", count)
	}
	
	// Add a player
	user1 := &server.User{IsSpectator: false}
	room.Connections[user1] = true
	
	if count := room.GetPlayersInRoom(); count != 1 {
		t.Errorf("Expected 1 player, got %d", count)
	}
	
	// Add a spectator
	user2 := &server.User{IsSpectator: true}
	room.Connections[user2] = true
	
	if count := room.GetPlayersInRoom(); count != 1 {
		t.Errorf("Expected 1 player (with 1 spectator), got %d", count)
	}
	
	// Add another player
	user3 := &server.User{IsSpectator: false}
	room.Connections[user3] = true
	
	if count := room.GetPlayersInRoom(); count != 2 {
		t.Errorf("Expected 2 players, got %d", count)
	}
}

func TestRoomInitPlayer(t *testing.T) {
	room := server.MakeRoom(1)
	
	// Test adding first player
	user1 := &server.User{IsSpectator: false}
	err := room.InitPlayer(user1)
	
	if err != nil {
		t.Errorf("Failed to add first player: %v", err)
	}
	
	if len(room.ReadyPlayers) != 1 {
		t.Errorf("Expected 1 ready player, got %d", len(room.ReadyPlayers))
	}
	
	// Test adding second player
	user2 := &server.User{IsSpectator: false}
	err = room.InitPlayer(user2)
	
	if err != nil {
		t.Errorf("Failed to add second player: %v", err)
	}
	
	if len(room.ReadyPlayers) != 2 {
		t.Errorf("Expected 2 ready players, got %d", len(room.ReadyPlayers))
	}
	
	// Test adding third player (should fail)
	user3 := &server.User{IsSpectator: false}
	err = room.InitPlayer(user3)
	
	if err == nil {
		t.Error("Expected error when adding third player, got nil")
	}
}

func TestRoomCheckAllReady(t *testing.T) {
	room := server.MakeRoom(1)
	
	// Test empty room
	if room.CheckAllReady() {
		t.Error("Empty room should not be ready")
	}
	
	// Add one player
	user1 := &server.User{IsSpectator: false}
	room.InitPlayer(user1)
	
	if room.CheckAllReady() {
		t.Error("Room with one player should not be ready")
	}
	
	// Add second player
	user2 := &server.User{IsSpectator: false}
	room.InitPlayer(user2)
	
	if !room.CheckAllReady() {
		t.Error("Room with two players should be ready")
	}
}

func TestRoomRemoveFromRoom(t *testing.T) {
	room := server.MakeRoom(1)
	
	// Add a player
	user := &server.User{IsSpectator: false}
	room.Connections[user] = true
	room.InitPlayer(user)
	
	// Remove the player
	room.RemoveFromRoom(user)
	
	if room.Connections[user] {
		t.Error("User should be marked as inactive after removal")
	}
	
	// Test removing from empty room
	server.MakeRoom(2).RemoveFromRoom(user) // Should not panic
} 
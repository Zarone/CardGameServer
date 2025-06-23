package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/Zarone/CardGameServer/cmd/gamemanager"
	"github.com/gorilla/websocket"
)

type User struct {
	Conn *websocket.Conn
	IsSpectator bool
}

type CoinFlip uint8
const (
	CoinFlipUnset = CoinFlip(0)
	CoinFlipHead  = CoinFlip(1)
	CoinFlipTail  = CoinFlip(2)
)

type RoomDescription string
const (
  DESC_FINISHED_INITIALIZATION  = RoomDescription("Finished Initialization...")
  DESC_PARAMETERS_READ          = RoomDescription("All players had parameters read...")
  DESC_HEADS_OR_TAILS_CHOSEN    = RoomDescription("Heads/Tails Chosen...")
  DESC_INITIAL_STATE_TO_CLIENT  = RoomDescription("Initial Game State Sent to Clients...")
	DESC_JUST_CREATED							= RoomDescription("Just Created...")
)

const PlayersToStartGame uint8 = 2

type Room struct {
	Connections             map[*User]bool
	PlayerToGamePlayerID    map[*User]uint8
	Game                    *gamemanager.Game
	ReadyPlayersMutex       sync.Mutex
	ReadyPlayers            []*User
	barrier                 *Barrier
	ExpectingCoinFlip       CoinFlip
	RoomNumber              uint8
	RoomDescription         RoomDescription
}

func MakeRoom(roomNumber uint8, cardHandler *gamemanager.CardHandler) *Room {
	ret := &Room{
		PlayerToGamePlayerID: make(map[*User]uint8),
		Connections: make(map[*User]bool),
		Game: gamemanager.MakeGame(cardHandler),
		ReadyPlayers: make([]*User, 0),
		ExpectingCoinFlip: CoinFlipUnset,
		RoomNumber: roomNumber,
		RoomDescription: DESC_JUST_CREATED,
		barrier: NewBarrier(int(PlayersToStartGame)),
	}
	return ret
}

// GetPlayersInRoom returns the number of players in the room,
// excluding spectators
func (r *Room) GetPlayersInRoom() uint8 {
	var num uint8 = 0

	for player, isActive := range r.Connections {
		if !player.IsSpectator && isActive {
			num++
		}
	}

	return num
}

func (r *Room) InitPlayer(user *User) error {
	r.ReadyPlayersMutex.Lock()
  defer r.ReadyPlayersMutex.Unlock()

	if len(r.ReadyPlayers) >= int(PlayersToStartGame) {
		return errors.New("too many players")
	}

	r.PlayerToGamePlayerID[user] = r.Game.AddPlayer() 
	r.ReadyPlayers = append(r.ReadyPlayers, user)
	

	return nil
}

func (r *Room) String() string {
	str := ""
	for user, isPresent := range r.Connections {
		str += fmt.Sprintf("[ConnectionPointer: %p, isPresent: %t, isSpectator: %t], ", user.Conn, isPresent, user.IsSpectator)
	}
	return str
}

// Sends the deck list to the game state manager to 
// set it up. Returns the gameID list in the same
// order as the cardID list
func (r *Room) initGameData(u *User, deck []uint) {
  r.Game.SetupPlayer(r.PlayerToGamePlayerID[u], deck)
}

func (r *Room) getInitData(u *User) Message[SetupResponse] {
  myDeck, oppDeck := r.Game.GetSetupData(r.PlayerToGamePlayerID[u])
	return Message[SetupResponse]{
    Content: SetupResponse{
      MyDeck: *myDeck,
      OppDeck: *oppDeck,
    },
    MessageType: gamemanager.MessageTypeSetup,
    Timestamp: timestamp(),
  }
}

// Takes info from client regarding their deck list 
// and such
func (r *Room) readSetupParams(user *User) (*Message[SetupContent], error) {
	if (user.IsSpectator) { return nil, nil }

	// if this is the last player to join
	// then tell the other players the game
	// is ready to setup
	defer r.wait(DESC_PARAMETERS_READ)

	// read in a message
	_, p, err := user.Conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error Reading Message {%s}", err)
	}
	
	var params Message[SetupContent] 
	if err := json.Unmarshal(p, &params); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err)
	} else if (params.MessageType != gamemanager.MessageTypeSetup) {
		return nil, fmt.Errorf("error setting up, message type is %v", params.MessageType)
	}

	return &params, nil
}

// Attempts to remove connection to the room specified by the request
func (r *Room) RemoveFromRoom(user *User) error {
	if len(r.Connections) == 0 {
		return errors.New("Error removing from room")
	}

	if (!user.IsSpectator) {
		fmt.Println("Implement player disconnect")
	}

	r.Connections[user] = false
  return nil
}

func (r *Room) readForActions(ws *websocket.Conn) (gamemanager.Action, error) {
	// read in a message
	_, p, err := ws.ReadMessage()
	if err != nil {
		return gamemanager.Action{}, fmt.Errorf("Error Reading Message {%s}", err)
	}

	// print out that message for clarity
	log.Printf("Message: %s", string(p))

  // unmarshal into gameaction
  var action Message[gamemanager.Action];
  err = json.Unmarshal(p, &action)
  if err != nil {
    return gamemanager.Action{}, fmt.Errorf("Error Getting Game Action from Message")
  }

	return action.Content, nil
}

func (r *Room) sendUpdateInfo(user *User, info *gamemanager.UpdateInfo) error {
  err := user.Conn.WriteJSON(
    Message[gamemanager.UpdateInfo]{
      Timestamp: timestamp(), 
      Content: *info, 
      MessageType: gamemanager.MessageTypeGameplay,
    },
  );
  if err != nil {
    return fmt.Errorf("Error writing message %s", err)
  }
  return nil
}

func (r *Room) processAction(user *User, action *gamemanager.Action) (*gamemanager.UpdateInfo, *gamemanager.UpdateInfo, error) {
	return r.Game.ProcessAction(r.PlayerToGamePlayerID[user], action)
}

func (r *Room) spectatorLoop(user *User) {
	for {
		_, err := r.readForActions(user.Conn)
		if (err != nil) {
			log.Println("Stopped reading from user, endcode: ", err)
			break
		}
	}
}

// returns (true, nil) if player 1 is going first
func (r *Room) askTurnOrder() (bool, error) {
	isHeads := rand.Intn(2) == 1 

	var userChoosingFlip *User
	var userWaiting *User
	if (isHeads == (r.ExpectingCoinFlip == CoinFlipHead)) {
		// if player 1 was right, then player 1 chooses
		userChoosingFlip = r.ReadyPlayers[0]
		userWaiting = r.ReadyPlayers[1]
	} else {
		// if player 1 was wrong, then player 2 chooses
		userChoosingFlip = r.ReadyPlayers[1]
		userWaiting = r.ReadyPlayers[0]
	}

	userChoosingFlip.Conn.WriteJSON(Message[StartGameContent]{
		Content: StartGameContent {
			IsChoosingTurnOrder: true,
		},
		MessageType: gamemanager.MessageTypeHeadsOrTails,
		Timestamp: timestamp(),
	})
	userWaiting.Conn.WriteJSON(Message[StartGameContent]{
		Content: StartGameContent {
			IsChoosingTurnOrder: false,
		},
		MessageType: gamemanager.MessageTypeFirstOrSecond,
		Timestamp: timestamp(),
	})

	var decisionResponse Message[StartGameContentChoice]

	_, p, err := userChoosingFlip.Conn.ReadMessage()
	if err != nil {
		return false, fmt.Errorf("error asking turn order: %s", err.Error())
	}

	err = json.Unmarshal(p, &decisionResponse)
	if err != nil {
		return false, fmt.Errorf("error asking turn order: %s", err.Error())
	}

	if decisionResponse.MessageType != gamemanager.MessageTypeFirstOrSecondChoice {
		return false, fmt.Errorf(
			"client response was expected to be a first or second choice, but was instead %d",
			decisionResponse.MessageType,
		)
	}
	
	return (userChoosingFlip == r.ReadyPlayers[0]) == decisionResponse.Content.First, nil
}

func (r *Room) headsOrTails(user *User) error {
	if (r.PlayerToGamePlayerID[user] == 0) {
		// this player chooses heads or tails
		var update Message[CoinFlipContent] = Message[CoinFlipContent]{
			Content: CoinFlipContent {
				IsChoosingFlip: true,
			},
			MessageType: gamemanager.MessageTypeHeadsOrTails,
			Timestamp: timestamp(),
		}
		err := user.Conn.WriteJSON(update)
		if err != nil {
			return fmt.Errorf("failed to WriteJSON for update: %s", err.Error())
		}
		var decisionResponse Message[CoinFlipContentChoice]

		_, p, err := user.Conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to ReadJSON for CoinFlipContentChoice: %s", err.Error())
		}

		err = json.Unmarshal(p, &decisionResponse)
		if err != nil {
			return errors.New("failed to decode JSON")
		}

		if decisionResponse.MessageType != gamemanager.MessageTypeCoinChoice {
			return fmt.Errorf(
				"client response was expected to be a coin choice, but was instead %v", 
				decisionResponse.MessageType,
			)
		}

		if decisionResponse.Content.Heads {
			r.ExpectingCoinFlip = CoinFlipHead
		} else {
			r.ExpectingCoinFlip = CoinFlipTail
		}
    fmt.Println("Set coin flip")
	} else if (r.PlayerToGamePlayerID[user] == 1) {
		// this player waits
		var update Message[CoinFlipContent] = Message[CoinFlipContent]{
			Content: CoinFlipContent {
				IsChoosingFlip: false,
			},
			MessageType: gamemanager.MessageTypeHeadsOrTails,
			Timestamp: timestamp(),
		}
		user.Conn.WriteJSON(update)
	} else {
		log.Println("Haven't handled scenario with more than 2 players")
	}

	return nil
}

func (r *Room) sendInitialGameState(goingFirst bool) {
	p1Info, p2Info := r.Game.StartGame(goingFirst)

	r.ReadyPlayers[0].Conn.WriteJSON(Message[gamemanager.UpdateInfo]{
		Content: *p1Info,
		MessageType: gamemanager.MessageTypeGameplay,
		Timestamp: timestamp(),
	})

	r.ReadyPlayers[1].Conn.WriteJSON(Message[gamemanager.UpdateInfo]{
		Content: *p2Info,
		MessageType: gamemanager.MessageTypeGameplay,
		Timestamp: timestamp(),
	})
}

func (r *Room) startGame(user *User) error {
	err := r.headsOrTails(user)
	if err != nil {
		return fmt.Errorf("error with heads or tails: %s", err.Error())
	}

	r.wait(DESC_HEADS_OR_TAILS_CHOSEN)

	if (r.ExpectingCoinFlip == CoinFlipUnset) {
		return errors.New("coin flip isn't set by evaluation time")
	}
	
	// arbitrarily let player 1 execute the following below,
	// or rather let the server call initiated by player 1 
	// execute the below code
	if r.PlayerToGamePlayerID[user] != 0 { 
		r.wait(DESC_INITIAL_STATE_TO_CLIENT)
		return nil
	} 


	goingFirst, err := r.askTurnOrder()
	if err != nil {
		return err
	}

	r.sendInitialGameState(goingFirst)

	r.wait(DESC_INITIAL_STATE_TO_CLIENT)

	return nil
}

func (r *Room) playerLoop(user *User) {
	for {
		action, err := r.readForActions(user.Conn)

		if (err != nil) {
			log.Println("Stopped reading from user, endcode: ", err)
			break
		}

    info, oppInfo, err := r.processAction(user, &action)
    if err != nil {
			log.Println("Error processing game action: ", err)
			break
    }

    err = r.sendUpdateInfo(user, info)
    if err != nil {
			log.Println("Stopped sending to user, endcode: ", err)
			break
    }

    id := r.PlayerToGamePlayerID[user]
    err = r.sendUpdateInfo(r.ReadyPlayers[1-id], oppInfo)
    if err != nil {
			log.Println("Stopped sending to user, endcode: ", err)
			break
    }

	}
}

func (r *Room) wait(newDescription RoomDescription) {
	fmt.Println("starting wait for:", newDescription)
	r.barrier.Wait()
	r.RoomDescription = newDescription
	fmt.Println("End wait for:", newDescription)
}

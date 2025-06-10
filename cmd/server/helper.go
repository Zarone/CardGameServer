package server

import (
  "fmt"
  "log"
  "strconv"
  "net/http"
)

func requestToRoomNumber(req *http.Request) uint8 {
  roomString := req.URL.Query().Get("room")

  var roomNum uint8
  if (roomString == "") {
    roomNum = 255
    log.Printf("No room selected, using default room (%d)\n", roomNum)
  } else {
    parsedRoomString, err := strconv.ParseInt(roomString, 10, 8)
    if err != nil {
      fmt.Printf("Error reading room number %s", roomString);
    }
    roomNum = uint8(parsedRoomString)
    log.Printf("Selected room: %s\n", roomString)
  }

  return roomNum
}


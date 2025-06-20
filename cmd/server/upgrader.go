package server

import (
  "net/http"
  "github.com/gorilla/websocket"
)

func originHandler(r *http.Request) bool {
  return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     originHandler,
}


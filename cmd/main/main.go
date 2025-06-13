package main

import (
  "fmt"
  "log"
  "net/http"
  "github.com/Zarone/CardGameServer/cmd/server"
)

func main() {
  myServer := server.MakeServer(&server.ServerSettings{}) 

  // example path: /socket?room=3&spectator=true
  http.HandleFunc("/socket", myServer.HandleWS)
  
  // Add handlers for rooms page and API
  http.HandleFunc("/", myServer.HandleRoomsPage)
  http.HandleFunc("/api/rooms", myServer.HandleRoomsAPI)

  fmt.Println("Hello from Server")
  log.Fatal(http.ListenAndServe(":3000", nil))
}

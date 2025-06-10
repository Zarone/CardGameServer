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

  fmt.Print("Hello from Server\n")
  log.Fatal(http.ListenAndServe(":3000", nil))
}

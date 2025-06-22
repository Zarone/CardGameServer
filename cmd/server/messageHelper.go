package server

import (
	"fmt"
  "time"
)

func timestamp() string {
	return fmt.Sprint(time.Now().Format("20060102150405"))
}

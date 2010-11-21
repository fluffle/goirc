package main

import (
	"fmt"
	"irc/server/config"
)

func main() {
	cfg := config.LoadConfig("server.cfg")
	for e, v := range(cfg.Errors) {
		fmt.Println(e, v)
	}
}

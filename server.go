package main

import (
	"fmt"
	"irc/server/config"
)

func main() {
	cfg := config.NewConfig()
	cfg.Ports[6667] = config.DefaultPort()
	cfg.Ports[6667].Port = 6667
	// cfg.Ports[6697] = &config.cPort{Port: 6697, SSL: true}
	fmt.Println(cfg.String())
	/*
	cfg, err := config.ConfigFromFile("server.cfg")
	if err != nil {
		fmt.Println(err)
		return
	}
	*/
	for i, p := range cfg.Ports {
		fmt.Printf("port %d\n", i)
		fmt.Println(p.String())
	}
}

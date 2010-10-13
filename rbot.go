package main

import (
	"irc"
	"fmt"
)

func main() {
	server := "irc.rizon.net"
	channel := "#vn-meta"
	ssl := false

	c := irc.New("raylu[BOT]", "rayluBOT", "rayluBOT")
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			conn.Join(channel)
		})

	for {
		fmt.Printf("Connecting to %s...\n", server)
		if err := c.Connect(server, ssl, ""); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			break
		}
		fmt.Println("Connected!")
		for err := range c.Err {
			fmt.Printf("goirc error: %s\n", err)
		}
	}
}

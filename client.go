package main

import (
	irc "github.com/fluffle/goirc/client"
	"fmt"
	"os"
	"bufio"
	"strings"
)

func main() {
	// create new IRC connection
	c := irc.New("GoTest", "gotest", "GoBot")
	c.Debug = true
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) { conn.Join("#go-nuts") })

	// connect to server
	if err := c.Connect("irc.freenode.net"); err != nil {
		fmt.Printf("Connection error: %s\n", err)
		return
	}

	// set up a goroutine to read commands from stdin
	in := make(chan string, 4)
	reallyquit := false
	go func() {
		con := bufio.NewReader(os.Stdin)
		for {
			s, err := con.ReadString('\n')
			if err != nil {
				// wha?, maybe ctrl-D...
				close(in)
				break
			}
			// no point in sending empty lines down the channel
			if len(s) > 2 {
				in <- s[0:len(s)-1]
			}
		}
	}()

	// set up a goroutine to do parsey things with the stuff from stdin
	go func() {
		for cmd := range in {
			if cmd[0] == ':' {
				switch idx := strings.Index(cmd, " "); {
				case cmd[1] == 'd':
					fmt.Printf(c.String())
				case cmd[1] == 'f':
					if len(cmd) > 2 && cmd[2] == 'e' {
						// enable flooding
						c.Flood = true
					} else if len(cmd) > 2 && cmd[2] == 'd' {
						// disable flooding
						c.Flood = false
					}
					for i := 0; i < 20; i++ {
						c.Privmsg("#", "flood test!")
					}
				case idx == -1:
					continue
				case cmd[1] == 'q':
					reallyquit = true
					c.Quit(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'j':
					c.Join(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'p':
					c.Part(cmd[idx+1 : len(cmd)])
				}
			} else {
				c.Raw(cmd)
			}
		}
	}()

	// stall here waiting for asplode on error channel
	for {
		for err := range c.Err {
			fmt.Printf("goirc error: %s\n", err)
		}
		if reallyquit {
			break
		}
		fmt.Println("Reconnecting...")
		if err := c.Connect("irc.freenode.net"); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			break
		}
	}
}

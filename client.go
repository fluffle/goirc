package main

import (
	"bufio"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var host *string = flag.String("host", "irc.freenode.net", "IRC server")
var channel *string = flag.String("channel", "#go-nuts", "IRC channel")

func main() {
	flag.Parse()

	// create new IRC connection
	c := irc.SimpleClient("GoTest", "gotest")
	c.EnableStateTracking()
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join(*channel) })

	// Set up a handler to notify of disconnect events.
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// Set up some simple commands, !bark and !roll.
	// The !roll command will also get the  "!help roll" command also.
	c.SimpleCommandFunc("bark", func(conn *irc.Conn, line *irc.Line) { conn.Privmsg(line.Target(), "Woof Woof") })
	c.SimpleCommandHelpFunc("roll", `Rolls a d6, "roll <n>" to roll n dice at once.`, func(conn *irc.Conn, line *irc.Line) {
		count := 1
		fields := strings.Fields(line.Message())
		if len(fields) > 1 {
			var err error
			if count, err = strconv.Atoi(fields[len(fields)-1]); err != nil {
				count = 1
			}
		}
		total := 0
		for i := 0; i < count; i++ {
			total += rand.Intn(6) + 1
		}
		conn.Privmsg(line.Target(), fmt.Sprintf("%d", total))
	})

	// Set up some commands that are triggered by a regex in a message.
	// It is important to see that UrlRegex could actually respond to some
	// of the Url's that YouTubeRegex listens to, because of this we put the
	// YouTube command at a higher priority, this way it will take precedence.
	c.CommandFunc(irc.YouTubeRegex, irc.YouTubeFunc, 10)
	c.CommandFunc(irc.UrlRegex, irc.UrlFunc, 0)

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
				reallyquit = true
				c.Quit("")
				break
			}
			// no point in sending empty lines down the channel
			if len(s) > 2 {
				in <- s[0 : len(s)-1]
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
						c.Config().Flood = true
					} else if len(cmd) > 2 && cmd[2] == 'd' {
						// disable flooding
						c.Config().Flood = false
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

	for !reallyquit {
		// connect to server
		if err := c.ConnectTo(*host); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}
		// wait on quit channel
		<-quit
	}
}

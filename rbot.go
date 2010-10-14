package main

import (
	"irc"
	"fmt"
	"os"
	"github.com/kless/goconfig/config"
)

var nick, server, user, trigger string
var ssl bool
var channels []string

func main() {
	parseConfig("rbot.conf")

	c := irc.New(nick, user, user)
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Println("Connected!")
			for _, c := range channels {
				conn.Join(c)
			}
		})
	c.AddHandler("privmsg", handlePrivmsg)

	for {
		fmt.Printf("Connecting to %s...\n", server)
		if err := c.Connect(server, ssl, ""); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			break
		}
		for err := range c.Err {
			fmt.Printf("goirc error: %s\n", err)
		}
	}
}

func parseConfig(confFile string) {
	conf, err := config.ReadDefault(confFile)
	if (err != nil) {
		fmt.Printf("Config error: %s\n", err); os.Exit(1)
	}

	server, err = conf.String("DEFAULT", "server")
	if err != nil { fmt.Printf("Config error: %s\n", err); os.Exit(1) }

	nick, err = conf.String("DEFAULT", "nick")
	if err != nil { fmt.Printf("Config error: %s\n", err); os.Exit(1) }

	user, err = conf.String("DEFAULT", "user")
	if err != nil { fmt.Printf("Config error: %s\n", err); os.Exit(1) }

	ssl, err = conf.Bool("DEFAULT", "ssl")
	if err != nil { fmt.Printf("Config error: %s\n", err); os.Exit(1) }

	sections := conf.Sections()
	channels = make([]string, len(sections)-1)
	i := 0
	for _, s := range sections {
		if (s != "DEFAULT") {
			channels[i] = s
			i++
		}
	}

	trigger, err = conf.String("DEFAULT", "trigger")
	if err != nil { fmt.Printf("Config error: %s\n", err); os.Exit(1) }
}

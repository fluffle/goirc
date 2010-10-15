package main

import (
	"irc"
	"fmt"
	"os"
	"strings"
	"github.com/kless/goconfig/config"
)

var trigger string
var sections []string

func main() {
	conf, err := config.ReadDefault("rbot.conf")
	if (err != nil) {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

	trigger = readConfString(conf, "DEFAULT", "trigger")

	sections = conf.Sections()
	for _, s := range sections {
		if strings.Index(s, " ") == -1 && s != "DEFAULT" {
			// found a network
			go connect(conf, s)
		}
	}

	<- make(chan bool)
}

func connect(conf *config.Config, network string) {
	if !readConfBool(conf, network, "autoconnect") {
		return
	}
	server := readConfString(conf, network, "server")
	nick := readConfString(conf, network, "nick")
	user := readConfString(conf, network, "user")
	ssl := readConfBool(conf, network, "ssl")

	c := irc.New(nick, user, user)
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Printf("Connected to %s!\n", conn.Host)
			for _, s := range sections {
				split := strings.Split(s, " ", 2)
				if len(split) == 2 && split[0] == network {
					// found a channel
					if readConfBool(conf, s, "autojoin") {
						fmt.Printf("Joining %s on %s\n", split[1], network)
						conn.Join(split[1])
					}
				}
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

func readConfString(conf *config.Config, section, option string) string {
	value, err := conf.String(section, option)
	if err != nil {
		panic(fmt.Sprintf("Config error: %s", err));
	}
	return value
}
func readConfBool(conf *config.Config, section, option string) bool {
	value, err := conf.Bool(section, option)
	if err != nil {
		panic(fmt.Sprintf("Config error: %s", err));
	}
	return value
}

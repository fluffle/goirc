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
var conf *config.Config

func main() {
	var err os.Error;
	conf, err = config.ReadDefault("rbot.conf")
	if (err != nil) {
		fmt.Printf("Config error: %s\n", err)
		os.Exit(1)
	}

	trigger = readConfString("DEFAULT", "trigger")
	readAuth()

	sections = conf.Sections()
	for _, s := range sections {
		if strings.Index(s, " ") == -1 && s != "DEFAULT" {
			// found a network
			go connect(s)
		}
	}

	<- make(chan bool)
}

func connect(network string) {
	if !readConfBool(network, "autoconnect") {
		return
	}
	server := readConfString(network, "server")
	nick := readConfString(network, "nick")
	user := readConfString(network, "user")
	ssl := readConfBool(network, "ssl")
	nickserv, _ := conf.String(network, "nickserv")

	c := irc.New(nick, user, user)
	c.Network = network
	c.AddHandler("connected",
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Printf("Connected to %s!\n", conn.Host)

			if len(nickserv) > 0 {
				conn.Privmsg("NickServ", "IDENTIFY " + nickserv)
			} else {
				autojoin(conn)
			}
		})
	c.AddHandler("privmsg", handlePrivmsg)
	c.AddHandler("mode", handleMode)

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

func autojoin(conn *irc.Conn) {
	for _, s := range sections {
		split := strings.Split(s, " ", 2)
		if len(split) == 2 && split[0] == conn.Network {
			// found a channel
			if readConfBool(s, "autojoin") {
				fmt.Printf("Joining %s on %s\n", split[1], conn.Network)
				conn.Join(split[1])
			}
		}
	}
}

func readConfString(section, option string) string {
	value, err := conf.String(section, option)
	if err != nil {
		panic(fmt.Sprintf("Config error: %s", err));
	}
	return value
}
func readConfBool(section, option string) bool {
	value, err := conf.Bool(section, option)
	if err != nil {
		panic(fmt.Sprintf("Config error: %s", err));
	}
	return value
}

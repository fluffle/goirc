package main

import (
	"fmt"
	"os"
	"strings"
	"irc"
	"github.com/kless/goconfig/config"
)

const authFile = "auth.conf"
var auth *config.Config

func readAuth() {
	var err os.Error;
	auth, err = config.ReadDefault(authFile)
	if (err != nil) {
		panic(fmt.Sprintf("Auth config error: %s", err))
	}
}

func addAccess(conn *irc.Conn, channel, nick, flags string) (string, string) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", ""
	}

	section := conn.Network + " " + channel
	cflags, _ := auth.String(section, n.Host)

	nflags := cflags
	for _, flag := range flags {
		if strings.IndexRune(cflags, flag) > -1 {
			// already has the flag
			continue
		}
		nflags += string(flag)
	}

	auth.AddOption(section, n.Host, nflags)
	if err := auth.WriteFile(authFile, 0644, ""); err != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}
	// config.WriteFile destroys the config, so
	readAuth()

	return n.Host, nflags
}

// passing a flag of "" will check if the user has any access
func hasAccess(conn *irc.Conn, channel, nick, flag string) bool {
	n := conn.GetNick(nick)
	if n == nil {
		return false
	}

	if owner, _ := auth.String(conn.Network, "owner"); owner == n.Host {
		return true
	}

	flags, err := auth.String(conn.Network + " " + channel, n.Host)
	if err != nil {
		return false
	}
	if strings.Index(flags, flag) > -1 {
		return true
	}
	return false
}

package main

import (
	"fmt"
	"os"
	"strings"
	irc "github.com/fluffle/goirc/client"
	"goconfig"
)

const authFile = "auth.conf"
var auth *config.Config
var ignores map[string]map[string]bool //network -> host -> true

func readAuth() {
	var err os.Error
	auth, err = config.ReadDefault(authFile)
	if err != nil {
		panic(fmt.Sprintf("Auth config error: %s", err))
	}

	if ignores == nil {
		ignores = make(map[string]map[string]bool)
		// parse ignores
		for _, s := range auth.Sections() {
			if strings.Index(s ," ") > -1 {
				// this is a channel
				continue
			}
			hosts, err := auth.Options(s)
			if err != nil {
				panic(fmt.Sprintf("Ignore config error in %s: %s", s, err))
			}
			if ignores[s] == nil {
				ignores[s] = make(map[string]bool)
			}
			for _, h := range hosts {
				if h == "owner" {
					continue
				}
				ignores[s][h] = true
			}
		}
	}
}

func user(nick *irc.Nick) string {
	if nick.Ident == "" || nick.Host == "" {
		return ""
	}
	if nick.Ident[0] == '~' {
		return nick.Ident[1:] + "@" + nick.Host
	}
	return nick.Ident + "@" + nick.Host
}

func addAccess(conn *irc.Conn, channel, nick, flags string) (string, string) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", ""
	}

	section := conn.Network + " " + channel
	user := user(n)
	cflags, _ := auth.String(section, user)

	nflags := cflags
	for _, flag := range flags {
		if strings.IndexRune(cflags, flag) > -1 {
			// already has the flag
			continue
		}
		nflags += string(flag)
	}

	auth.AddOption(section, user, nflags)
	if updateAuth() != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}

	return user, nflags
}
func removeAccess(conn *irc.Conn, channel, nick, flags string) (string, string) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", ""
	}

	section := conn.Network + " " + channel
	user := user(n)
	cflags, _ := auth.String(section, user)

	nflags := ""
	for _, flag := range cflags {
		if strings.IndexRune(flags, flag) < 0 {
			// we're not removing this flag
			nflags += string(flag)
		}
	}

	auth.AddOption(section, user, nflags)
	if updateAuth() != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}

	return user, nflags
}
func removeUser(conn *irc.Conn, channel, nick string) (string, bool) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", false
	}

	section := conn.Network + " " + channel
	user := user(n)

	if !auth.RemoveOption(section, user) {
		return user, false
	}
	if updateAuth() != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}

	return user, true
}

func hasAccess(conn *irc.Conn, nick *irc.Nick, channel, flag string) bool {
	user := user(nick)
	if owner, _ := auth.String(conn.Network, "owner"); owner == user {
		return true
	}

	flags, err := auth.String(conn.Network + " " + channel, user)
	if err != nil {
		return false
	}
	if flag == "" || strings.IndexAny(flags, flag) > -1 {
		return true
	}
	return false
}

func addIgnore(conn *irc.Conn, channel string, n *irc.Nick) bool {
	if ignores[conn.Network][n.Host] {
		return false
	}
	ignores[conn.Network][n.Host] = true
	auth.AddOption(conn.Network, n.Host, "z")
	if updateAuth() != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}

	return true
}
func removeIgnore(conn *irc.Conn, channel, nick string) (string, bool) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", false
	}

	if _, ok := ignores[conn.Network][n.Host]; !ok {
		return n.Host, false
	}
	ignores[conn.Network][n.Host] = false, false
	auth.RemoveOption(conn.Network, n.Host)
	if updateAuth() != nil {
		say(conn, channel, "Error while writing to %s", authFile)
	}

	return n.Host, true
}

func updateAuth() os.Error {
	if err := auth.WriteFile(authFile, 0644, ""); err != nil {
		return err
	}
	// config.WriteFile destroys the config, so
	readAuth()
	return nil
}

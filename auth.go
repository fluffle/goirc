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

func user(nick *irc.Nick) string {
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

// this allows target to be a channel or a privmsg in which the channel is the first argument
// passing a flag of "" will check if the user has any access
// returns the channel the user has access on and remaining args
// or a blank channel if the user doesn't have access on that channel
func hasAccess(conn *irc.Conn, nick, target, args, flag string) (string, string) {
	n := conn.GetNick(nick)
	if n == nil {
		return "", args
	}

	// figure out what the channel and args are
	var channel string
	if isChannel(target) {
		channel = target
	} else {
		split := strings.Split(args, " ", 2)
		if isChannel(split[0]) {
			channel = split[0]
			if len(split) == 2 {
				args = split[1]
			} else {
				args = ""
			}
		} else {
			return "", args
		}
	}

	// actually check access
	user := user(n)
	if owner, _ := auth.String(conn.Network, "owner"); owner == user {
		return channel, args
	}

	flags, err := auth.String(conn.Network + " " + channel, user)
	if err != nil {
		return "", args
	}
	if strings.Index(flags, flag) > -1 {
		return channel, args
	}
	return "", args
}

func updateAuth() os.Error {
	if err := auth.WriteFile(authFile, 0644, ""); err != nil {
		return err
	}
	// config.WriteFile destroys the config, so
	readAuth()
	return nil
}

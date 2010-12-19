package main

import (
	irc "github.com/fluffle/goirc/client"
)

func nick(conn *irc.Conn, nick *irc.Nick, args, target string) {
	if len(args) == 0 {
		return
	}
	owner, _ := auth.String(conn.Network, "owner");
	if owner == user(nick) {
		conn.Nick(args)
	}
}

func csay(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "s")
	if len(channel) > 0 {
		say(conn, channel, args)
	}
}

package main

import (
	"irc"
	//"fmt"
	"strings"
)

var commands = map [string]func(*irc.Conn, string, string, string) {
	"kill": kill,
}

func handlePrivmsg(conn *irc.Conn, line *irc.Line) {
	target := line.Args[0]
	if target[0] == '#' || target[0] == '&' {
		// message to a channel
		command(conn, line.Nick, line.Text, target)
	} else if target == nick {
		// message to us
		command(conn, line.Nick, line.Text, target)
	}
}

func command(conn *irc.Conn, nick, text, target string) {
	if text[:len(trigger)] != trigger {
		return
	}
	split := strings.Split(text, " ", 2)
	if len(split[0]) < 2 {
		return
	}
	handler := commands[split[0][1:]]
	if handler != nil {
		if len(split) > 1 {
			handler(conn, nick, split[1], target)
		} else {
			handler(conn, nick, "", target)
		}
	}
}

func kill(conn *irc.Conn, nick, args, target string) {
	conn.Action(target, "dies.")
}

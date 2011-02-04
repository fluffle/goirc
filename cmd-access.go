package main

import (
	irc "github.com/fluffle/goirc/client"
	"strings"
)

func add(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "a")
	if channel == "" {
		return
	}
	split := strings.Fields(args)
	if len(split) != 2 {
		return
	}
	host, nflags := addAccess(conn, channel, split[0], strings.TrimSpace(split[1]))
	if host == "" {
		say(conn, target, "Could not find nick %s", split[0])
	} else {
		say(conn, target, "%s now has flags %s", host, nflags)
	}
}
func remove(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "a")
	if channel == "" {
		return
	}
	split := strings.Fields(args)

	if len(split) == 2 {
		host, nflags := removeAccess(conn, channel, split[0], strings.TrimSpace(split[1]))
		if host == "" {
			say(conn, target, "Could not find nick %s", split[0])
		} else {
			say(conn, target, "%s now has flags %s", host, nflags)
		}
	} else if len(split) == 1 {
		host, removed := removeUser(conn, channel, split[0])
		if host == "" {
			say(conn, target, "Could not find nick %s", split[0])
		} else if removed {
			say(conn, target, "Removed %s", host)
		} else {
			say(conn, target, "%s did not have any flags", host)
		}
	}
}

func flags(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "")
	if channel == "" {
		return
	}

	query := strings.TrimSpace(args)
	if query == "" {
		query = nick.Nick
	}
	n := conn.GetNick(query)
	if n == nil {
		say(conn, target, "Could not find nick %s", query)
		return
	}

	user := user(n)
	if owner, _ := auth.String(conn.Network, "owner"); owner == user {
		say(conn, target, "%s is the owner", user)
		return
	}

	flags, _ := auth.String(conn.Network + " " + channel, user)
	if flags == "" {
		say(conn, target, "%s has no flags", user)
	} else {
		say(conn, target, "%s: %s", user, flags)
	}
}

func ignore(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "a")
	if channel == "" {
		return
	}

	n := conn.GetNick(strings.TrimSpace(args))
	if n == nil {
		say(conn, target, "Could not find nick %s", args)
		return
	}
	if nick == n {
		say(conn, target, "%s: you cannot ignore yourself", nick.Nick)
	}
	owner, _ := auth.String(conn.Network, "owner")
	if owner == user(n) {
		return
	}

	if addIgnore(conn, channel, n) {
		say(conn, target, "Ignoring %s", n.Host)
	} else {
		say(conn, target, "%s is already ignored", n.Host)
	}
}
func unignore(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "a")
	if channel == "" {
		return
	}
	host, success := removeIgnore(conn, channel, strings.TrimSpace(args))
	if host == "" {
		say(conn, target, "Could not find nick %s", args)
	} else if !success {
		say(conn, target, "%s was not ignored", host)
	} else {
		say(conn, target, "Unignored %s", host)
	}
}

func accesslist(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := parseAccess(conn, nick, target, args, "")
	if channel == "" {
		return
	}

	owner, err := auth.String(conn.Network, "owner")
	if err == nil && strings.Contains(owner, args)  {
		say(conn, nick.Nick, "%s is the owner", owner)
	}

	for i, _ := range(ignores[conn.Network]) {
		if strings.Contains(i, args) {
			say(conn, nick.Nick, "%s is being ignored", i)
		}
	}

	section := conn.Network + " " + channel
	users, err := auth.Options(section)
	if err != nil {
		say(conn, nick.Nick, "%s: Error while getting users")
		return
	}
	for _, u := range users {
		if strings.Contains(u, args) {
			flags, err := auth.String(section, u)
			if err == nil {
				say(conn, nick.Nick, "%s: %s", u, flags)
			}
		}
	}
}

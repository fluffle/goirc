package main

import (
	irc "github.com/fluffle/goirc/client"
	"fmt"
	"strings"
	"runtime"
	"reflect"
	"http"
	"xml"
	"strconv"
)

var commands = map [string]func(*irc.Conn, *irc.Nick, string, string) {
	// access
	"flags": flags,
	"add": add,
	"remove": remove,
	"ignore": ignore,
	"unignore": unignore,
	"list": accesslist,

	// admin
	"nick": nick,
	"say": csay,

	// op
	"halfop": halfop,
	"hop": halfop,
	"op": op,
	"deop": deop,
	"dehalfop": dehalfop,
	"dehop": dehalfop,
	"voice": voice,
	"devoice": devoice,
	"kick": kick,
	"k": kick,
	"b": ban,
	"ban": ban,
	"unban": unban,
	"u": unban,
	"kb": kickban,
	"topic": topic,
	"appendtopic": appendtopic,
	"part": part,
	"ops": highlightOps,

	// google
	"tr": translate,
	"roman": roman,
	"calc": calc,
}

func handlePrivmsg(conn *irc.Conn, line *irc.Line) {
	nick := conn.GetNick(line.Nick)
	if nick == nil {
		return
	}
	nick.Host = line.Host
	if ignores[conn.Network][nick.Host] {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from", r)
			callers := make([]uintptr, 10)
			runtime.Callers(4, callers)
			cutoff := runtime.FuncForPC(reflect.ValueOf(handlePrivmsg).Pointer()).Entry()
			for _, pc := range callers {
				function := runtime.FuncForPC(pc - 1)
				if function.Entry() == cutoff {
					break
				}
				file, line := function.FileLine(pc - 1)
				fmt.Printf("%s:%d\n\t%s\n", file, line, function.Name())
			}
		}
	}()

	target := line.Args[0]
	if isChannel(target) {
		// message to a channel
		if !command(conn, nick, line.Args[1], target) {
			var video string
			if start := strings.Index(line.Args[1], "youtube.com/watch?v="); start > -1 {
				video = line.Args[1][start+20:]
			}
			if start := strings.Index(line.Args[1], "youtu.be/"); start > -1 {
				video = line.Args[1][start+9:]
			}
			if video != "" {
				if end := strings.IndexAny(video, " &#"); end > -1 {
					video = video[0:end]
				}
				youtube(conn, nick, video, target)
			}
		}
	} else if target == conn.Me.Nick {
		// message to us
		command(conn, nick, line.Args[1], line.Nick)
	}
}

func handleMode(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] == conn.Me.Nick && line.Args[1] == "+r" {
		autojoin(conn)
	}
}

func handleJoin(conn *irc.Conn, line *irc.Line) {
	// autovoice users with v flag
	if line.Nick == conn.Me.Nick {
		return
	}

	channel := conn.GetChannel(line.Args[0])
	if channel == nil || !channel.Modes.Moderated {
		return
	}

	privs := conn.Me.Channels[channel]
	if !(privs.Op || privs.Admin || privs.HalfOp || privs.Owner) {
		return
	}
	nick := conn.GetNick(line.Nick)
	if nick == nil {
		return
	}
	if hasAccess(conn, nick, line.Nick, "v") {
		conn.Mode(line.Args[0], "+v " + line.Nick)
	}
}

func handleInvite(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] != conn.Me.Nick {
		return
	}

	user := line.Src[strings.Index(line.Src, "!")+1:]
	if user[0] == '~' {
		user = user[1:]
	}

	owner, _ := auth.String(conn.Network, "owner")
	if user == owner {
		conn.Join(line.Args[1])
	}
}

func isChannel(target string) bool {
	return target[0] == '#' || target[0] == '&'
}

func command(conn *irc.Conn, nick *irc.Nick, text, target string) bool {
	if !strings.HasPrefix(text, trigger) {
		return false
	}
	split := strings.Split(text, " ", 2)
	if len(split[0]) < 2 {
		return false
	}
	handler := commands[split[0][1:]]
	if handler != nil {
		if len(split) > 1 {
			handler(conn, nick, split[1], target)
		} else {
			handler(conn, nick, "", target)
		}
		return true
	}
	return false
}

func say(conn *irc.Conn, target, message string, a ...interface{}) {
	if len(a) > 0 {
		message = fmt.Sprintf(message, a...)
	}
	text := strings.Replace(message, "\n", " ", -1)
	if isChannel(target) {
		conn.Privmsg(target, text)
	} else {
		conn.Notice(target, text)
	}
}

func youtube(conn *irc.Conn, nick *irc.Nick, video, channel string) {
	url := fmt.Sprintf("http://gdata.youtube.com/feeds/api/videos/%s?v=2", video)
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		return
	}

	type duration struct {
		Seconds string "attr"
	}
	type group struct {
		Duration duration
	}
	type entry struct {
		Title string
		Group group
	}
	var yte = entry{"", group{duration{""}}}

	err = xml.Unmarshal(response.Body, &yte)
	if err != nil {
		return
	}

	seconds, err := strconv.Atoui(yte.Group.Duration.Seconds)
	if err == nil {
		minutes := seconds / 60
		seconds = seconds % 60
		hours := minutes / 60
		minutes = minutes % 60
		var durationStr string
		if hours > 0 {
			durationStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			durationStr = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
		say(conn, channel, "%s's video: %s, %s", nick.Nick, yte.Title, durationStr)
	} else {
		say(conn, channel, "%s's video: %s", nick.Nick, yte.Title)
	}
}

// this allows target to be a channel or a privmsg in which the channel is the first argument
// passing a flag of "" will check if the user has any access
// returns the channel the user has access on and remaining args
// or a blank channel if the user doesn't have access on that channel
func parseAccess(conn *irc.Conn, nick *irc.Nick, target, args, flag string) (string, string) {
	channel, args := parseChannel(target, args)
	if channel == "" {
		return "", args
	}
	if hasAccess(conn, nick, channel, flag) {
		return channel, args
	}
	return "", args
}

func parseChannel(target, args string) (string, string) {
	var channel string
	if isChannel(target) {
		channel = target
	} else {
		split := strings.Split(args, " ", 2)
		if split[0] != "" && isChannel(split[0]) {
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
	return channel, args
}

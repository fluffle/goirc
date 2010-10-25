package main

import (
	"irc"
	"fmt"
	"strings"
	"http"
	"io/ioutil"
	"json"
	"container/vector"
	"html"
	"xml"
	"strconv"
)

var commands = map [string]func(*irc.Conn, *irc.Nick, string, string) {
	"tr": translate,
	"flags": flags,
	"add": add,
	"remove": remove,
	"topic": topic,
	"appendtopic": appendtopic,
	"say": csay,
}

const googleAPIKey = "ABQIAAAA6-N_jl4ETgtMf2M52JJ_WRQjQjNunkAJHIhTdFoxe8Di7fkkYhRRcys7ZxNbH3MIy_MKKcEO4-9_Ag"

func handlePrivmsg(conn *irc.Conn, line *irc.Line) {
	nick := conn.GetNick(line.Nick)
	if nick == nil {
		return
	}
	nick.Host = line.Host

	target := line.Args[0]
	if isChannel(target) {
		// message to a channel
		var video string
		if strings.HasPrefix(line.Text, "http://www.youtube.com/watch?v=") {
			video = line.Text[31:]
		} else if strings.HasPrefix(line.Text, "http://www.youtube.com/watch?v=") {
			video = line.Text[27:]
		}
		if video != "" {
			if amp := strings.Index(video, "&"); amp > -1 {
				video = video[0:amp]
			}
			if pound := strings.Index(video, "#"); pound > -1 {
				video = video[0:pound]
			}
			youtube(conn, nick, video, target)
		} else {
			command(conn, nick, line.Text, target)
		}
	} else if target == conn.Me.Nick {
		// message to us
		command(conn, nick, line.Text, line.Nick)
	}
}

func handleMode(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] == conn.Me.Nick && line.Text == "+r" {
		autojoin(conn)
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
		conn.Join(line.Text)
	}
}

func isChannel(target string) bool {
	return target[0] == '#' || target[0] == '&'
}

func command(conn *irc.Conn, nick *irc.Nick, text, target string) {
	if !strings.HasPrefix(text, trigger) {
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

func say(conn *irc.Conn, target, message string, a ...interface{}) {
	text := strings.Replace(fmt.Sprintf(message, a...), "\n", " ", -1)
	if isChannel(target) {
		conn.Privmsg(target, text)
	} else {
		conn.Notice(target, text)
	}
}

func youtube(conn *irc.Conn, nick *irc.Nick, video, channel string) {
	url := fmt.Sprintf("http://gdata.youtube.com/feeds/api/videos/%s?v=2", video)
	response, _, err := http.Get(url)
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

func translate(conn *irc.Conn, nick *irc.Nick, args, target string) {
	var langPairs vector.StringVector
	for {
		split := strings.Split(args, " ", 2)
		if len(split) == 2 && len(split[0]) == 5 && split[0][2] == '|' {
			langPairs.Push("&langpair=" + split[0])
			args = split[1]
		} else {
			break
		}
	}

	var url string
	if langPairs.Len() > 0 {
		// translate
		langPairsSlice := []string(langPairs)
		url = fmt.Sprintf("http://ajax.googleapis.com/ajax/services/language/translate?v=1.0&q=%s%s&key=%s",
		                   http.URLEscape(args), strings.Join(langPairsSlice, ""), googleAPIKey)
	} else {
		// language detect
		url = fmt.Sprintf("http://ajax.googleapis.com/ajax/services/language/detect?v=1.0&q=%s&key=%s",
		                   http.URLEscape(args), googleAPIKey)
	}

	response, _, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		say(conn, target, "%s: Error while requesting translation", nick.Nick); return
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		say(conn, target, "%s: Error while downloading translation", nick.Nick); return
	}

	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		say(conn, target, "%s: Error while decoding translation", nick.Nick); return
	}
	if result["responseStatus"] != float64(200) {
		say(conn, target, "%s: %s", nick.Nick, result["responseDetails"]); return
	}

	if langPairs.Len() > 0 {
		// translate
		sayTr(conn, target, result["responseData"])
	} else {
		// language detect
		var data map[string]interface{} = result["responseData"].(map[string]interface{})
		say(conn, target, "Language: %s, confidence: %f, is reliable: %t", data["language"], data["confidence"], data["isReliable"])
	}
}

func sayTr(conn *irc.Conn, target string, data interface{}) {
	switch t := data.(type) {
	case []interface{}:
		var dataList []interface{} = data.([]interface{})
		for _, d := range dataList {
			var innerData map[string]interface{} = d.(map[string]interface{})
			sayTr(conn, target, innerData["responseData"])
		}
	case map[string]interface{}:
		trText := data.(map[string]interface{})["translatedText"].(string)
		say(conn, target, html.UnescapeString(trText))
	}
}

func add(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := hasAccess(conn, nick, target, args, "a")
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
	channel, args := hasAccess(conn, nick, target, args, "a")
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
	channel, args := hasAccess(conn, nick, target, args, "")
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

func topic(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := hasAccess(conn, nick, target, args, "t")
	if channel == "" {
		return
	}
	section := conn.Network + " " + channel
	if args != "" {
		updateConf(section, "basetopic", args)
		conn.Topic(channel, args)
	} else {
		basetopic, _ := conf.String(section, "basetopic")
		say(conn, nick.Nick, "Basetopic: %s", basetopic)
	}
}
func appendtopic(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := hasAccess(conn, nick, target, args, "t")
	if channel == "" {
		return
	}
	c := conn.GetChannel(channel)
	if c == nil {
		say(conn, target, "Error while getting channel information for %s", channel)
		return
	}

	section := conn.Network + " " + channel
	basetopic, _ := conf.String(section, "basetopic")
	if basetopic == "" || !strings.HasPrefix(c.Topic, basetopic) {
		basetopic = c.Topic
		say(conn, nick.Nick, "New basetopic: %s", basetopic)
		updateConf(section, "basetopic", basetopic)
	}
	conn.Topic(channel, basetopic + args)
}

func csay(conn *irc.Conn, nick *irc.Nick, args, target string) {
	channel, args := hasAccess(conn, nick, target, args, "s")
	if channel != "" {
		say(conn, channel, args)
	}
}

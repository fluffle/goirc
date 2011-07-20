package main

import (
	"http"
	"strings"
	"json"
	"bufio"
	"io/ioutil"
	"time"
	irc "github.com/fluffle/goirc/client"
)

var justinTv map[string]bool = make(map[string]bool)

func announceJustinTv(conn *irc.Conn, channel string) {
	for {
		time.Sleep(30000000000) // 30 seconds
		live := pollJustinTv()
		if len(live) > 0 {
			say(conn, channel, "Now streaming: %s; http://stream.sc2gg.com/", strings.Join(live, ", "))
		}
	}
}

func pollJustinTv() []string {
	newStreams := make(map[string]bool)
	qChannel := ""
	streams, err := http.Get("http://stream.sc2gg.com/streams")
	if err == nil {
		defer streams.Body.Close()
		b := bufio.NewReader(streams.Body)
		for {
			line, err := b.ReadString('\n')
			if err != nil {
				break
			}
			username := strings.Split(line, "\t", 2)[1]
			if len(username) > 3 && username[:3] == "/j/" {
				username = username[3:len(username)-1]
				newStreams[username] = false
				qChannel += username + ","
			}
		}
	}

	if len(qChannel) == 0 {
		return nil
	}
	response, err := http.Get("http://api.justin.tv/api/stream/list.json?channel=" + qChannel[:len(qChannel)-1])
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	var result []interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil
	}
	live := make([]string, 0)
	for _, r := range(result) {
		data := r.(map[string]interface{})
		channel := data["channel"].(map[string]interface{})
		username := channel["login"].(string)
		newStreams[username] = true
		if !justinTv[username] {
			live = append(live, username)
		}
	}
	justinTv = newStreams
	return live
}

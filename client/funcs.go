package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type youTubeVideo struct {
	Entry struct {
		Info struct {
			Title struct {
				Text string `json:"$t"`
			} `json:"media$title"`
			Description struct {
				Text string `json:"$t"`
			} `json:"media$description"`
		} `json:"media$group"`
		Rating struct {
			Likes    string `json:"numLikes"`
			Dislikes string `json:"numDislikes"`
		} `json:"yt$rating"`
		Statistics struct {
			Views string `json:"viewCount"`
		} `json:"yt$statistics"`
	} `json:entry`
}

func UrlFunc(conn *Conn, line *Line) {
	text := line.Message()
	if regex, err := regexp.Compile(`(\s|^)(http://|https://)(.*?)(\s|$)`); err == nil {
		url := strings.TrimSpace(regex.FindString(text))
		if url != "" {
			if resp, err := http.Get(url); err == nil {
				if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
					defer resp.Body.Close()
					if content, err := ioutil.ReadAll(resp.Body); err == nil {
						if regex, err := regexp.Compile(`<title>(.*?)</title>`); err == nil {
							if regex.Match([]byte(content)) {
								conn.Privmsg(line.Target(), strings.TrimSpace(regex.FindStringSubmatch(string(content))[1]))
							}
						}
					}
				}
			}
		}
	}
}

func YouTubeFunc(conn *Conn, line *Line) {
	text := line.Message()
	if regex, err := regexp.Compile(`(\s|^)(http://|https://)?(www.)?(youtube.com/watch\?v=|youtu.be/)(.*?)(\s|$|\&|#)`); err == nil {
		if regex.Match([]byte(text)) {
			matches := regex.FindStringSubmatch(text)
			id := matches[len(matches)-2]
			url := fmt.Sprintf("https://gdata.youtube.com/feeds/api/videos/%s?v=2&alt=json", id)
			if resp, err := http.Get(url); err == nil {
				defer resp.Body.Close()
				if contents, err := ioutil.ReadAll(resp.Body); err == nil {
					var data youTubeVideo
					if err := json.Unmarshal(contents, &data); err == nil {
						conn.Privmsg(line.Target(), fmt.Sprintf("%s - %s views (%s likes, %s dislikes)", data.Entry.Info.Title.Text, data.Entry.Statistics.Views, data.Entry.Rating.Likes, data.Entry.Rating.Dislikes))
					}
				}
			}
		}
	}
}

package main

import (
	"irc"
	"fmt"
	"os"
	"strings"
	"http"
	"json"
	"io/ioutil"
	"container/vector"
	"html"
	"regexp"
	"strconv"
)

const googleAPIKey = "ABQIAAAA6-N_jl4ETgtMf2M52JJ_WRQjQjNunkAJHIhTdFoxe8Di7fkkYhRRcys7ZxNbH3MIy_MKKcEO4-9_Ag"

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

func calc(conn *irc.Conn, nick *irc.Nick, args, target string) {
	if args == "" {
		return
	}
	url := fmt.Sprintf("http://www.google.com/ig/calculator?hl=en&q=%s&key=%s", http.URLEscape(args), googleAPIKey)
	response, _, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		say(conn, target, "%s: Error while requesting calculation", nick.Nick); return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		say(conn, target, "%s: Error while downloading calculation", nick.Nick); return
	}

	re := regexp.MustCompile(`{lhs: "(.*)",rhs: "(.*)",error: "(.*)",icc: (true|false)}`)
	result := re.FindSubmatch(b)
	if len(result) != 5 || len(result[3]) > 1 {
		say(conn, target, "%s: Error while calculating.", nick.Nick); return
	}

	str := fmt.Sprintf("%s = %s", result[1], result[2])
	output := ""
	// decode unicode escapes
	for str != "" {
		var err os.Error
		var rune int
		rune, _, str, err = strconv.UnquoteChar(str, 0)
		if err != nil {
			say(conn, target, "%s: Error while decoding.", nick.Nick); return
		}
		output += string(rune)
	}
	output = html.UnescapeString(output)
	say(conn, target, output)
}

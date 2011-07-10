package main

import (
	irc "github.com/fluffle/goirc/client"
	"fmt"
	"strings"
	"os"
	"http"
	"net"
	"io/ioutil"
	"container/vector"
	"html"
	"json"
	"regexp"
	"strconv"
	"utf8"
)

const googleAPIKey = "ABQIAAAA6-N_jl4ETgtMf2M52JJ_WRQjQjNunkAJHIhTdFoxe8Di7fkkYhRRcys7ZxNbH3MIy_MKKcEO4-9_Ag"

func translate(conn *irc.Conn, nick *irc.Nick, args, target string) {
	var langPairs vector.StringVector
	for {
		field := strings.IndexAny(args, " 　") // handle spaces and ideographic spaces (U+3000)
		if field != -1 {
			first := args[:field]
			if len(first) == 5 && first[2] == '|' {
				langPairs.Push("&langpair=" + first)
				if args[field] == ' ' {
					args = args[field+1:]
				} else {
					args = args[field+utf8.RuneLen(3000):]
				}
			} else {
				break
			}
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

	response, err := http.Get(url)
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

func roman(conn *irc.Conn, nick *irc.Nick, args, target string) {
	if args == "" {
		return
	}

	var sourcelang, targetlang string
	if utf8.NewString(args).IsASCII() {
		sourcelang = "en"
	} else {
		sourcelang = "ja"
	}
	targetlang, _ = conf.String(conn.Network, "roman")
	if targetlang == "" {
		targetlang = "ja"
	}
	url := fmt.Sprintf("http://translate.google.com/translate_a/t?client=t&hl=%s&sl=%s&tl=en-US&text=%s",
		targetlang, sourcelang, http.URLEscape(args))

	b, err := getUserAgent(url)
	if err != nil {
		say(conn, target, "%s: Error while requesting romanization", nick.Nick); return
	}

	result := strings.Split(string(b), `"`, 7)
	if len(result) < 7 {
		say(conn, target, "%s: Error while parsing romanization", nick.Nick); return
	}

	if (sourcelang == "en" && !strings.Contains(args, " ")) {
		// Google duplicates when there is only one source word
		source := utf8.NewString(result[1])
		romanized := utf8.NewString(result[5])
		source_left := source.Slice(0, source.RuneCount()/2)
		source_right := source.Slice(source.RuneCount()/2, source.RuneCount())
		romanized_left := romanized.Slice(0, romanized.RuneCount()/2)
		romanized_right :=romanized.Slice(romanized.RuneCount()/2, romanized.RuneCount())
		if (source_left == source_right &&
			strings.ToLower(romanized_left) == strings.ToLower(romanized_right)) {
			say(conn, target, "%s: %s", source_left, romanized_left)
			return
		}
	}
	say(conn, target, "%s: %s", result[1], result[5])
}

func calc(conn *irc.Conn, nick *irc.Nick, args, target string) {
	if args == "" {
		return
	}
	url := fmt.Sprintf("http://www.google.com/ig/calculator?hl=en&q=%s&key=%s", http.URLEscape(args), googleAPIKey)

	b, err := getUserAgent(url)
	if err != nil {
		say(conn, target, "%s: Error while requesting calculation", nick.Nick); return
	}

	re := regexp.MustCompile(`{lhs: "(.*)",rhs: "(.*)",error: "(.*)",icc: (true|false)}`)
	result := re.FindSubmatch(b)
	if len(result) != 5 {
		say(conn, target, "%s: Error while parsing.", nick.Nick)
		return
	}
	if len(result[3]) > 1 {
		output := fmt.Sprintf(`"%s"`, result[3])
		error := parseCalc(output)
		if error != "" {
			say(conn, target, "%s: Error: %s", nick.Nick, error)
		} else {
			say(conn, target, "%s: Error while calculating and error while decoding error.", nick.Nick)
		}
		return
	}
	if len(result[1]) == 0 || len(result[2]) == 0 {
		say(conn, target, "%s: Error while calculating.", nick.Nick)
		return
	}

	output := fmt.Sprintf(`"%s = %s"`, result[1], result[2])
	output = parseCalc(output)
	if output == "" {
		say(conn, target, "%s: Error while decoding.", nick.Nick); return
	}
	say(conn, target, output)
}

func parseCalc(output string) string {
	parsed, err := strconv.Unquote(output)
	if err != nil {
		return ""
	}
	parsed = html.UnescapeString(parsed)
	parsed = strings.Replace(parsed, "<sup>", "^(", -1)
	parsed = strings.Replace(parsed, "</sup>", ")", -1)
	return parsed
}

// make a GET request with a fake user agent
// this is definitely not for those undocumented Google APIs
func getUserAgent(urlstr string) ([]byte, os.Error) {
	url, err := http.ParseURL(urlstr)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", url.Host + ":http")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		return nil, err
	}
	req.UserAgent = "Mozilla/5.0"

	httpconn := http.NewClientConn(conn, nil)
	response, err := httpconn.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	return b, err
}

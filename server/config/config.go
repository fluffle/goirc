package config

import (
	"io"
	"os"
	"fmt"
	"strconv"
	"scanner"
)

type Config struct {
	fn string
	scan *scanner.Scanner

	// Ports we listen on.
	Ports map[int]*cPort
	// People with teh p0wer.
	Opers map[string]*cOper
	// Servers we link to on the network.
	Links map[string]*cLink
	// Servers/nickmasks/IPs that are unwanted.
	Bans []*cBan

	// Server info (name, admins, etc.)
	Info *cInfo

	// Server settings
	Settings *cSettings

	// Parse errors
	Errors []os.Error
}

type configMap  map[string]func(*Config)
type keywordMap map[string]func(*Config, interface{})

var configKeywords = configMap{
	"port": (*Config).parsePort,
//	"oper": (*Config).parseOper,
//	"link": (*Config).parseLink,
//	"ban":  (*Config).parseBan,
//	"info": (*Config).parseInfo,
//	"set":  (*Config).parseSettings,
}

func LoadConfig(filename string) *Config {
	conf := &Config{fn: filename}
	conf.initialise()
	if fh, err := os.Open(conf.fn, os.O_RDONLY, 0644); err == nil {
		conf.Parse(fh)
		fh.Close()
	} else {
		conf.Errors = append(conf.Errors, err)
	}
	return conf
}

func (conf *Config) initialise() {
	conf.Ports = make(map[int]*cPort)
	conf.Opers = make(map[string]*cOper)
	conf.Links = make(map[string]*cLink)
	conf.Bans = make([]*cBan, 0)
	conf.Info = &cInfo{}
	conf.Settings = &cSettings{}
	conf.Errors = make([]os.Error, 0)
}

func (conf *Config) Rehash() {
	neu := LoadConfig(conf.fn)
	if len(neu.Errors) > 0 {
		conf.Errors = neu.Errors
	} else {
		conf = neu
	}
}

func (conf *Config) Parse(io io.Reader) {
	s := &scanner.Scanner{}
	s.Init(io)
	s.Filename = conf.fn
	conf.scan = s
	tok, text := conf.next()
	for tok != scanner.EOF {
		// This external loop should only parse Config things
		if f, ok := configKeywords[text]; ok {
			f(conf)
		} else {
			conf.parseError("Invalid top-level keyword '%s'", text)
		}
		fmt.Printf("Token: '%s', type %s\n", s.TokenText(), scanner.TokenString(tok))
		tok, text = conf.next()
	}
}

func (conf *Config) parseKwBlock(dst interface{}, bt string, kw keywordMap) {
	if ok := conf.expect("{"); !ok {
		conf.parseError("Expected %s configuration block.", bt)
		return
	}
	tok, text := conf.next()
	for tok != scanner.EOF {
		if f, ok := kw[text]; ok {
			if ok = conf.expect("="); ok {
				f(conf, dst)
			}
		} else if text == "}" {
			break
		} else {
			conf.parseError("Invalid %s keyword '%s'", bt, text)
		}
		tok, text = conf.next()
	}
}

var booleans = map[string]bool {
	"true": true,
	"yes": true,
	"on": true,
	"1": true,
	"false": false,
	"no": false,
	"off": false,
	"0": false,
}

func (conf *Config) expectBool() (bool, bool) {
	tok, text := conf.next()
	if val, ok := booleans[text]; tok == scanner.Ident && ok {
		return val, ok
	}
	conf.parseError("Expected boolean, got '%s'", text)
	return false, false
}

func (conf *Config) expectInt() (int, bool) {
	tok, text := conf.next()
	num, err := strconv.Atoi(text)
	if tok != scanner.Int || err != nil {
		conf.parseError("Expected integer, got '%s'", text)
		return 0, false
	}
	return num, true
}

func (conf *Config) expect(str string) bool {
	_, text := conf.next()
	if text != str {
		conf.parseError("Expected '%s', got '%s'", str, text)
		return false
	}
	return true
}

func (conf *Config) next() (int, string) {
	tok := conf.scan.Scan()
	text := conf.scan.TokenText()
	if tok == scanner.String {
		// drop "quotes" -> quotes
		text = text[1:len(text)-1]
	}
	return tok, text
}

func (conf *Config) parseError(err string, args ...interface{}) {
	err = conf.scan.Pos().String() + ": " + err
	conf.Errors = append(conf.Errors, os.NewError(fmt.Sprintf(err, args...)))
}

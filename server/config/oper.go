package config

import (
	"fmt"
	"strings"
	"scanner"
)

type cOper struct {
	Username, Password string
	HostMask []string

	// Permissions for oper
	CanKill, CanBan, CanRenick, CanLink  bool
}

var operKeywords = keywordMap{
	"password": (*Config).parseOperPassword,
	"hostmask": (*Config).parseOperHostMask,
	"kill": (*Config).parseOperKill,
	"ban": (*Config).parseOperBan,
	"renick": (*Config).parseOperRenick,
	"link": (*Config).parseOperLink,
}

func defaultOper() *cOper {
	return &cOper{
		HostMask: []string{},
		CanKill: true, CanBan: true,
		CanRenick: false, CanLink: false,
	}
}

func (o *cOper) String() string {
    str := []string{fmt.Sprintf("oper \"%s\" {", o.Username)}
	str = append(str, fmt.Sprintf("\tpassword = \"%s\"", o.Password))
	if len(o.HostMask) == 0 {
		str = append(str, fmt.Sprintf("\thostmask = \"*@*\""))
	} else {
		for _, h := range o.HostMask {
			str = append(str, fmt.Sprintf("\thostmask = \"%s\"", h))
		}
	}
	str = append(str,
		fmt.Sprintf("\tkill    = %t", o.CanKill),
		fmt.Sprintf("\tban     = %t", o.CanBan),
		fmt.Sprintf("\trenick  = %t", o.CanRenick),
		fmt.Sprintf("\tlink    = %t", o.CanLink),
		"}",
	)
	return strings.Join(str, "\n")
}

func (conf *Config) parseOper() {
	oper := defaultOper()
	tok, text := conf.next()
	if tok != scanner.String && tok != scanner.Ident {
		conf.parseError("Invalid username '%s'", text)
		return
	}
	oper.Username = text
	conf.parseKwBlock(oper, "oper", operKeywords)
	fmt.Println(oper.String())
}

func (conf *Config) parseOperPassword(oi interface{}) {
	oper := oi.(*cOper)
	if pass, ok := conf.expectString(); ok {
		oper.Password = pass
	}
}

func (conf *Config) parseOperHostMask(oi interface{}) {
	oper := oi.(*cOper)
	if mask, ok := conf.expectString(); ok {
		oper.HostMask = append(oper.HostMask, mask)
	}
}

func (conf *Config) parseOperKill(oi interface{}) {
	oper := oi.(*cOper)
	if kill, ok := conf.expectBool(); ok {
		oper.CanKill = kill
	}
}

func (conf *Config) parseOperBan(oi interface{}) {
	oper := oi.(*cOper)
	if ban, ok := conf.expectBool(); ok {
		oper.CanBan = ban
	}
}

func (conf *Config) parseOperRenick(oi interface{}) {
	oper := oi.(*cOper)
	if renick, ok := conf.expectBool(); ok {
		oper.CanRenick = renick
	}
}

func (conf *Config) parseOperLink(oi interface{}) {
	oper := oi.(*cOper)
	if link, ok := conf.expectBool(); ok {
		oper.CanLink = link
	}
}

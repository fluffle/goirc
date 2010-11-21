package config

import (
	"fmt"
	"strings"
	"scanner"
)

type cPort struct {
	Port     int
	BindIP, Family, Class string

	// Is port a tls.Listener? Does it support compression (no)?
	SSL, Zip bool

	// address == "<BindIP>:<Port>"
	address	 string
}

var portKeywords = keywordMap{
//	"bind_ip": (*Config).parsePortBindIP,
//	"family": (*Config).parsePortFamily,
	"class": (*Config).parsePortClass,
	"ssl": (*Config).parsePortSSL,
//	"zip": (*Config).parsePortZip,
}

var cPortDefaults = cPort{
	BindIP: "", Family: "tcp", Class: "client",
	SSL: false, Zip: false,
}

func defaultPort() *cPort {
	p := cPortDefaults
	return &p
}

func (p *cPort) String() string {
    str := []string{fmt.Sprintf("port %d {", p.Port)}
	if p.BindIP != "" {
		str = append(str, "\tbind_ip = " + p.BindIP)
	}
	str = append(str,
		fmt.Sprintf("\tfamily = \"%s\"",p.Family),
		fmt.Sprintf("\tclass  = \"%s\"", p.Class),
		fmt.Sprintf("\tssl    = %t", p.SSL),
		fmt.Sprintf("\tzip    = %t", p.Zip),
		"}",
	)
	return strings.Join(str, "\n")
}

func (conf *Config) parsePort() {
	port := defaultPort()
	portnum, ok := conf.expectInt()
	if !ok || portnum > 65535 || portnum < 1024 {
		conf.parseError("Invalid port '%s'", portnum)
		port = nil
	} else {
		port.Port = portnum
		conf.Ports[portnum] = port
	}
	if conf.scan.Peek() != '\n' {
		conf.parseKwBlock(port, "port", portKeywords)
	}
	fmt.Println(port.String())
}

func (conf *Config) parsePortClass(pi interface{}) {
	port := pi.(*cPort)
	tok, text := conf.next()
	if tok == scanner.String && (text == "server" || text == "client") {
		port.Class = text
	} else {
		conf.parseError(
			"Port class must be \"server\" or \"client\", got '%s'", text)
	}
}

func (conf *Config) parsePortSSL(pi interface{}) {
	port := pi.(*cPort)
	if ssl, ok := conf.expectBool(); ok {
		port.SSL = ssl
	}
}


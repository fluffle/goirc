package config

import (
	"net"
	"fmt"
	"strings"
	"scanner"
)

type cPort struct {
	Port     int
	BindIP   net.IP // bind to a specific IP for listen port
	Class    string // "server" or "client"

	// Is port a tls.Listener? Does it support compression (no)?
	SSL, Zip bool
}

var portKeywords = keywordMap{
	"bind_ip": (*Config).parsePortBindIP,
	"class": (*Config).parsePortClass,
	"ssl": (*Config).parsePortSSL,
	"zip": (*Config).parsePortZip,
}

func defaultPort() *cPort {
	return &cPort{
		BindIP: nil, Class: "client",
		SSL: false, Zip: false,
	}
}

func (p *cPort) String() string {
    str := []string{fmt.Sprintf("port %d {", p.Port)}
	if p.BindIP != nil {
		str = append(str,
			fmt.Sprintf("\tbind_ip = \"%s\"", p.BindIP.String()))
	}
	str = append(str,
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

func (conf *Config) parsePortBindIP(pi interface{}) {
	port := pi.(*cPort)
	_, text := conf.next()
	if ip := net.ParseIP(text); ip != nil {
		port.BindIP = ip
	} else {
		conf.parseError("'%s' is not a valid IP address", text)
	}
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

func (conf *Config) parsePortZip(pi interface{}) {
	port := pi.(*cPort)
	if zip, ok := conf.expectBool(); ok {
		port.Zip = zip
	}
}

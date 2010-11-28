package config

import (
	"os"
	"fmt"
	"net"
	"strings"
)

type Config struct {
	fn string

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

/* Port configuration */
type cPort struct {
	Port     int
	BindIP   net.IP // bind to a specific IP for listen port
	Class    string // "server" or "client"

	// Is port a tls.Listener? Does it support compression (no)?
	SSL, Zip bool
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
			fmt.Sprintf("\tbind_ip = %s", p.BindIP.String()))
	}
	str = append(str,
		fmt.Sprintf("\tclass  = %s", p.Class),
		fmt.Sprintf("\tssl    = %t", p.SSL),
		fmt.Sprintf("\tzip    = %t", p.Zip),
		"}",
	)
	return strings.Join(str, "\n")
}

/* Oper configuration */
type cOper struct {
	Username, Password string
	HostMask []string

	// Permissions for oper
	CanKill, CanBan, CanRenick, CanLink  bool
}

func defaultOper() *cOper {
	return &cOper{
		HostMask: []string{},
		CanKill: true, CanBan: true,
		CanRenick: false, CanLink: false,
	}
}

func (o *cOper) String() string {
    str := []string{fmt.Sprintf("oper %s {", o.Username)}
	str = append(str, fmt.Sprintf("\tpassword = %s", o.Password))
	for _, h := range o.HostMask {
		str = append(str, fmt.Sprintf("\thostmask = %s", h))
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

/* Link configuration */
type cLink struct {
	Server      string // Server name for link
	Address     string // {ip,ip6,host}:port
	ReceivePass string // Password when server connects to us 
	ConnectPass string // Password when we connect to server

	// Do we use tls.Dial? or compression (no)? Do we auto-connect on start?
	SSL, Zip, Auto bool
}

/* Static ban configuration */
type cBan interface {
	Match(string) bool
	Reason() string
}

// G-Line etc; 
type cBanNick struct {
	NickMask string // nick!ident@host
	Reason   string
}

// Z-Line
type cBanIP struct {
	Address string // ip (or hostname), plus optional CIDR netmask
	Reason  string
	ip		string // parsed into these
	cidr	int
}

// CTCP version ban
type cBanVersion struct {
	VersionRegex string // regex to match against version reply
	Reason       string
}

// Ban server from linking to network
type cBanServer struct {
	ServerMask string // matched against name of linked server
	Reason     string
}

/* IRCd settings */
type cSettings struct {
	SSLKey, SSLCert, SSLCACert string
	MaxChans, MaxConnsPerIP int
	LogFile string
}

/* IRCd information */
type cInfo struct {
	Name, Network, Info, MOTDFile string
	Admins []string
	Numeric int
}


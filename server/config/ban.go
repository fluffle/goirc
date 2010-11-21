package config

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


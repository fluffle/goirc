package config

type cLink struct {
	Server      string // Server name for link
	Address     string // {ip,ip6,host}:port
	ReceivePass string // Password when server connects to us 
	ConnectPass string // Password when we connect to server

	// Do we use tls.Dial? or compression (no)? Do we auto-connect on start?
	SSL, Zip, Auto bool
}


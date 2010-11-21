package config

type cSettings struct {
	SSLKey, SSLCert, SSLCACert string
	MaxChans, MaxConnsPerIP int
	LogFile string
}


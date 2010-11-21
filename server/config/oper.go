package config

type cOper struct {
	Username, Password string
	HostMask []string

	// Permissions for oper
	CanKill, CanBan, CanNick, CanLink  bool
}

var cOperDefaults = &cOper{
	HostMask: []string{"*@*"},
	CanKill: true, CanBan: true,
	CanNick: false, CanLink: false,
}


package config

import (
	"io"
	"os"
	"fmt"
	"net"
	"strconv"
)

%% machine config;
%% write data;

// some helpers
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

func getbool(val []byte) bool {
	return booleans[string(val)]
}

func getint(val []byte) int {
	if v, err := strconv.Atoi(string(val)); err == nil {
		return v
	}
	return 0
}

%%{
###############################################################################
# basic parser bits
	# mark the current position for acquiring data
	action mark { mark = p }

	# acceptable boolean values
	true = "yes" | "on" | "1" | "true" ;
	false = "no" | "off" | "0" | "false" ;
	boolean = true | false ;

	# IP addresses, v4 and v6
	octet = digit         # 0-9
		| digit digit       # 00-99
		| [01] digit digit  # 000-199
		| "2" [0-4] digit   # 200-249
		| "2" "5" [0-5] ;   # 250-255
	ipv4addr = octet "." octet "." octet "." octet ;
	ipv6addr = ( xdigit{1,4} ":" ){7} xdigit{1,4}      # all 8 blocks
		| xdigit{0,4} "::"                               # special case
		| ( xdigit{1,4} ":" ){1,6} ":"                   # first 1-6, ::
		| ":" ( ":" xdigit{1,4} ){1,6}                   # ::, final 1-6
		| xdigit{1,4} ":" ( ":" xdigit{1,4} ){1,6}       # 1::1-6
		| ( xdigit{1,4} ":" ){2}( ":" xdigit{1,4} ){1,5} # 2::1-5
		| ( xdigit{1,4} ":" ){3}( ":" xdigit{1,4} ){1,4} # 3::1-4
		| ( xdigit{1,4} ":" ){4}( ":" xdigit{1,4} ){1,5} # 4::1-3
		| ( xdigit{1,4} ":" ){5}( ":" xdigit{1,4} ){1,2} # 5::1-2
		| ( xdigit{1,4} ":" ){6}( ":" xdigit{1,4} ) ;    # 6::1

	# acceptable password chars:
	# anything in the normal ascii set apart from spaces and control chars
	passchar = ascii -- ( space | cntrl ) ;

	# acceptable hostmask chars:
	# alphanumeric, plus "*", "?" and "."
	hostchar = ( alnum | "*" | "?" | "." ) ;

###############################################################################
# "port" configuration parser
	# Actions to create a cPort and save it into the Config struct
	action p_new { cur = defaultPort() }
	action p_save {
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}

	# parse and save the port number
	action px_port {
		cur.(*cPort).Port = getint(data[mark:p])
	}
	pv_port = digit+ >mark %px_port ;

	# parse a bind_ip statement and save the IP
	action px_bind_ip {
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	pv_bind_ip = (ipv4addr | ipv6addr) >mark %px_bind_ip ;
	ps_bind_ip = "bind_ip" " "+ "=" " "+ pv_bind_ip ;

	# parse a class statement and save it
	action px_class {
		cur.(*cPort).Class = string(data[mark:p])
	}
	pv_class = ( "server" | "client" ) >mark %px_class ;
	ps_class = "class" " "+ "=" " "+ pv_class ;

	# parse SSL and Zip booleans
	action px_ssl {
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	pv_ssl = boolean >mark %px_ssl ;
	ps_ssl = "ssl" " "+ "=" " "+ pv_ssl ;

	action px_zip {
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	pv_zip = boolean >mark %px_zip ;
	ps_zip = "zip" " "+ "=" " "+ pv_zip ;

	# a port statement can be any one of the above statements
	p_stmt = ( ps_bind_ip | ps_class | ps_ssl | ps_zip ) ;
	# and a port block combines one or more statements
	p_block = "{" space* ( p_stmt | (p_stmt " "* "\n" space*)+ ) space* "}" ;

	# a port configuration can either be:
	# port <port>\n
	# port <port> { <statement> ... }
	p_line = "port" >p_new " "+ pv_port " "* "\n" %p_save ;
	p_defn = "port" >p_new " "+ pv_port " "+ p_block %p_save ;
	port_config = space* ( p_line | p_defn ) space* ;


###############################################################################
# "oper" configuration parser
	# actions to create a new cOper and save it into the Config struct
	action o_new { cur = defaultOper() }
	action o_save {
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}

	# parse and save the username
	action ox_username {
		cur.(*cOper).Username = string(data[mark:p])
	}
	ov_username = alpha >mark alnum* %ox_username ;

	# parse a password statement and save it
	action ox_password {
		cur.(*cOper).Password = string(data[mark:p])
	}
	ov_password = passchar+ >mark %ox_password ;
	os_password = "password" " "+ "=" " "+ ov_password ;

	# parse a hostmask statement and save it
	action ox_hostmask {
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	ov_hostmask = hostchar+ >mark "@" hostchar+ %ox_hostmask ;
	os_hostmask = "hostmask" " "+ "=" " "+ ov_hostmask ;

	# parse and save the various oper permissions
	action ox_kill {
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	ov_kill = boolean >mark %ox_kill ;
	os_kill = "kill" " "+ "=" " "+ ov_kill ;

	action ox_ban {
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	ov_ban = boolean >mark %ox_ban ;
	os_ban = "ban" " "+ "=" " "+ ov_ban ;

	action ox_renick {
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	ov_renick = boolean >mark %ox_renick ;
	os_renick = "renick" " "+ "=" " "+ ov_renick ;

	action ox_link {
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	ov_link = boolean >mark %ox_link ;
	os_link = "link" " "+ "=" " "+ ov_link ;

	# an oper statement can be any of the above statements
	o_stmt = ( os_password
	  | os_hostmask
	  | os_kill
	  | os_ban
	  | os_renick
	  | os_link ) ;
	# and an oper block combines one or more statements
	o_block = "{" space* ( o_stmt | (o_stmt " "* "\n" space*)+ ) space* "}" ;

	# an oper configuration looks like:
	# oper <username> { <statement> ... }
	oper_config =  "oper" >o_new " "+ ov_username
					" "+ o_block %o_save space* ;

#	config = portconfig+
#	  | operconfig+
#	  | linkconfig*
#	  | infoconfig
#	  | settings ;

	main := ( port_config | oper_config )+ ;
}%%

func (conf *Config) Parse(r io.Reader) {
	cs, p, mark, pe, eof, buflen := 0, 0, 0, 0, 0, 16384
	done := false
	var cur interface{}
	data := make([]byte, buflen)

	%% write init;

	for !done {
		n, err := r.Read(data)
		pe = p + n
		if err == os.EOF {
			fmt.Println("yeahhhhh.")
			done = true
			eof = pe
		}

		%% write exec;

	}

	if cs < config_first_final {
		fmt.Printf("Parse error at %d near '%s'\n", p, data[p:p+10])
	}

	for _, port := range conf.Ports {
		fmt.Println(port.String())
	}
	for _, oper := range conf.Opers {
		fmt.Println(oper.String())
	}
}

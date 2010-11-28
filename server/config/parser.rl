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

%%{
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

	# Actions to create a cPort and save it into the Config struct
	action new_port { cur = defaultPort() }
	action save_port {
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}

	# parse and save the port number
	action set_portnum {
		cur.(*cPort).Port, _ = strconv.Atoi(string(data[mark:p]))
	}
	portnum = digit+ >mark %set_portnum ;

	# parse a bind_ip statement and save the IP
	action set_bindip {
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	bindip = (ipv4addr | ipv6addr) >mark %set_bindip ;
	portbindip = "bind_ip" " "+ "=" " "+ bindip ;

	# parse a class statement and save it
	action set_class {
		cur.(*cPort).Class = string(data[mark:p])
	}
	portclass = "class" " "+ "=" " "+ ("server" | "client" >mark %set_class) ;

	# parse SSL and Zip booleans
	action set_ssl {
		cur.(*cPort).SSL = booleans[string(data[mark:p])]
	}
	portssl = "ssl" " "+ "=" " "+ (boolean >mark %set_ssl) ;
	action set_zip {
		cur.(*cPort).Zip = booleans[string(data[mark:p])]
	}
	portzip = "zip" " "+ "=" " "+ (boolean >mark %set_zip) ;

	portstmt = ( portbindip | portclass | portssl | portzip ) ;
	portblock = "{" space* ( portstmt | (portstmt " "* "\n" space* )+ ) space* "}" ;

	# a port configuration can either be:
	# port <portnum>\n
	# port <portnum> { portblock }
	basicport = "port" >new_port " "+ portnum " "* "\n" %save_port;
	portdefn = "port" >new_port " "+ portnum " "+ portblock %save_port ;
	portconfig = space* ( basicport | portdefn ) space*;

#	config = portconfig+
#	  | operconfig+
#	  | linkconfig*
#	  | infoconfig
#	  | settings ;

	main := portconfig+;
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
}

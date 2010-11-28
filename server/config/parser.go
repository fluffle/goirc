
// line 1 "parser.rl"
package config

import (
	"io"
	"os"
	"fmt"
	"net"
	"strconv"
)


// line 12 "parser.rl"

// line 17 "parser.go"
var config_start int = 1
var config_first_final int = 560
var config_error int = 0

var config_en_main int = 1


// line 13 "parser.rl"

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


// line 208 "parser.rl"


func (conf *Config) Parse(r io.Reader) {
	cs, p, mark, pe, eof, buflen := 0, 0, 0, 0, 0, 16384
	done := false
	var cur interface{}
	data := make([]byte, buflen)

	
// line 61 "parser.go"
	cs = config_start

// line 217 "parser.rl"

	for !done {
		n, err := r.Read(data)
		pe = p + n
		if err == os.EOF {
			fmt.Println("yeahhhhh.")
			done = true
			eof = pe
		}

		
// line 76 "parser.go"
	{
	if p == pe { goto _test_eof }
	switch cs {
	case -666: // i am a hack D:
	fallthrough
case 1:
	switch data[p] {
		case 32: goto st2
		case 111: goto tr2
		case 112: goto tr3
	}
	if 9 <= data[p] && data[p] <= 13 { goto st2 }
	goto st0
st0:
cs = 0;
	goto _out;
st2:
	p++
	if p == pe { goto _test_eof2 }
	fallthrough
case 2:
	switch data[p] {
		case 32: goto st2
		case 112: goto tr3
	}
	if 9 <= data[p] && data[p] <= 13 { goto st2 }
	goto st0
tr3:
// line 77 "parser.rl"
	{ cur = defaultPort() }
	goto st3
tr633:
// line 78 "parser.rl"
	{
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}
// line 77 "parser.rl"
	{ cur = defaultPort() }
	goto st3
tr637:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
// line 77 "parser.rl"
	{ cur = defaultPort() }
	goto st3
st3:
	p++
	if p == pe { goto _test_eof3 }
	fallthrough
case 3:
// line 133 "parser.go"
	if data[p] == 111 { goto st4 }
	goto st0
st4:
	p++
	if p == pe { goto _test_eof4 }
	fallthrough
case 4:
	if data[p] == 114 { goto st5 }
	goto st0
st5:
	p++
	if p == pe { goto _test_eof5 }
	fallthrough
case 5:
	if data[p] == 116 { goto st6 }
	goto st0
st6:
	p++
	if p == pe { goto _test_eof6 }
	fallthrough
case 6:
	if data[p] == 32 { goto st7 }
	goto st0
st7:
	p++
	if p == pe { goto _test_eof7 }
	fallthrough
case 7:
	if data[p] == 32 { goto st7 }
	if 48 <= data[p] && data[p] <= 57 { goto tr8 }
	goto st0
tr8:
// line 41 "parser.rl"
	{ mark = p }
	goto st8
st8:
	p++
	if p == pe { goto _test_eof8 }
	fallthrough
case 8:
// line 174 "parser.go"
	switch data[p] {
		case 10: goto tr9
		case 32: goto tr10
	}
	if 48 <= data[p] && data[p] <= 57 { goto st8 }
	goto st0
tr9:
// line 85 "parser.rl"
	{
		cur.(*cPort).Port = getint(data[mark:p])
	}
	goto st560
tr323:
// line 91 "parser.rl"
	{
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	goto st560
tr584:
// line 98 "parser.rl"
	{
		cur.(*cPort).Class = string(data[mark:p])
	}
	goto st560
tr602:
// line 105 "parser.rl"
	{
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	goto st560
tr623:
// line 111 "parser.rl"
	{
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	goto st560
st560:
	p++
	if p == pe { goto _test_eof560 }
	fallthrough
case 560:
// line 216 "parser.go"
	switch data[p] {
		case 32: goto tr631
		case 111: goto tr632
		case 112: goto tr633
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr631 }
	goto st0
tr631:
// line 78 "parser.rl"
	{
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}
	goto st561
tr635:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
	goto st561
st561:
	p++
	if p == pe { goto _test_eof561 }
	fallthrough
case 561:
// line 245 "parser.go"
	switch data[p] {
		case 32: goto st561
		case 111: goto tr2
		case 112: goto tr3
	}
	if 9 <= data[p] && data[p] <= 13 { goto st561 }
	goto st0
tr2:
// line 133 "parser.rl"
	{ cur = defaultOper() }
	goto st9
tr632:
// line 78 "parser.rl"
	{
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}
// line 133 "parser.rl"
	{ cur = defaultOper() }
	goto st9
tr636:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
// line 133 "parser.rl"
	{ cur = defaultOper() }
	goto st9
st9:
	p++
	if p == pe { goto _test_eof9 }
	fallthrough
case 9:
// line 282 "parser.go"
	if data[p] == 112 { goto st10 }
	goto st0
st10:
	p++
	if p == pe { goto _test_eof10 }
	fallthrough
case 10:
	if data[p] == 101 { goto st11 }
	goto st0
st11:
	p++
	if p == pe { goto _test_eof11 }
	fallthrough
case 11:
	if data[p] == 114 { goto st12 }
	goto st0
st12:
	p++
	if p == pe { goto _test_eof12 }
	fallthrough
case 12:
	if data[p] == 32 { goto st13 }
	goto st0
st13:
	p++
	if p == pe { goto _test_eof13 }
	fallthrough
case 13:
	if data[p] == 32 { goto st13 }
	if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto tr16 }
	} else if data[p] >= 65 {
		goto tr16
	}
	goto st0
tr16:
// line 41 "parser.rl"
	{ mark = p }
	goto st14
st14:
	p++
	if p == pe { goto _test_eof14 }
	fallthrough
case 14:
// line 327 "parser.go"
	if data[p] == 32 { goto tr17 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st14 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st14 }
	} else {
		goto st14
	}
	goto st0
tr17:
// line 141 "parser.rl"
	{
		cur.(*cOper).Username = string(data[mark:p])
	}
	goto st15
st15:
	p++
	if p == pe { goto _test_eof15 }
	fallthrough
case 15:
// line 348 "parser.go"
	switch data[p] {
		case 32: goto st15
		case 123: goto st16
	}
	goto st0
st16:
	p++
	if p == pe { goto _test_eof16 }
	fallthrough
case 16:
	switch data[p] {
		case 32: goto st16
		case 98: goto st17
		case 104: goto st142
		case 107: goto st156
		case 108: goto st175
		case 112: goto st194
		case 114: goto st217
	}
	if 9 <= data[p] && data[p] <= 13 { goto st16 }
	goto st0
st17:
	p++
	if p == pe { goto _test_eof17 }
	fallthrough
case 17:
	if data[p] == 97 { goto st18 }
	goto st0
st18:
	p++
	if p == pe { goto _test_eof18 }
	fallthrough
case 18:
	if data[p] == 110 { goto st19 }
	goto st0
st19:
	p++
	if p == pe { goto _test_eof19 }
	fallthrough
case 19:
	if data[p] == 32 { goto st20 }
	goto st0
st20:
	p++
	if p == pe { goto _test_eof20 }
	fallthrough
case 20:
	switch data[p] {
		case 32: goto st20
		case 61: goto st21
	}
	goto st0
st21:
	p++
	if p == pe { goto _test_eof21 }
	fallthrough
case 21:
	if data[p] == 32 { goto st22 }
	goto st0
st22:
	p++
	if p == pe { goto _test_eof22 }
	fallthrough
case 22:
	switch data[p] {
		case 32: goto st22
		case 102: goto tr33
		case 110: goto tr34
		case 111: goto tr35
		case 116: goto tr36
		case 121: goto tr37
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr32 }
	goto st0
tr32:
// line 41 "parser.rl"
	{ mark = p }
	goto st23
st23:
	p++
	if p == pe { goto _test_eof23 }
	fallthrough
case 23:
// line 432 "parser.go"
	switch data[p] {
		case 10: goto tr39
		case 32: goto tr40
		case 125: goto tr41
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr38 }
	goto st0
tr38:
// line 168 "parser.rl"
	{
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	goto st24
tr188:
// line 154 "parser.rl"
	{
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	goto st24
tr203:
// line 162 "parser.rl"
	{
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	goto st24
tr225:
// line 180 "parser.rl"
	{
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	goto st24
tr246:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st24
tr274:
// line 174 "parser.rl"
	{
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	goto st24
st24:
	p++
	if p == pe { goto _test_eof24 }
	fallthrough
case 24:
// line 482 "parser.go"
	switch data[p] {
		case 32: goto st24
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st24 }
	goto st0
tr41:
// line 168 "parser.rl"
	{
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	goto st562
tr190:
// line 154 "parser.rl"
	{
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	goto st562
tr205:
// line 162 "parser.rl"
	{
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	goto st562
tr227:
// line 180 "parser.rl"
	{
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	goto st562
tr276:
// line 174 "parser.rl"
	{
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	goto st562
st562:
	p++
	if p == pe { goto _test_eof562 }
	fallthrough
case 562:
// line 525 "parser.go"
	switch data[p] {
		case 32: goto tr635
		case 111: goto tr636
		case 112: goto tr637
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr635 }
	goto st0
tr39:
// line 168 "parser.rl"
	{
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	goto st25
tr85:
// line 154 "parser.rl"
	{
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	goto st25
tr99:
// line 162 "parser.rl"
	{
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	goto st25
tr120:
// line 180 "parser.rl"
	{
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	goto st25
tr140:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st25
tr157:
// line 174 "parser.rl"
	{
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	goto st25
st25:
	p++
	if p == pe { goto _test_eof25 }
	fallthrough
case 25:
// line 575 "parser.go"
	switch data[p] {
		case 32: goto st25
		case 98: goto st26
		case 104: goto st45
		case 107: goto st59
		case 108: goto st78
		case 112: goto st97
		case 114: goto st109
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st25 }
	goto st0
st26:
	p++
	if p == pe { goto _test_eof26 }
	fallthrough
case 26:
	if data[p] == 97 { goto st27 }
	goto st0
st27:
	p++
	if p == pe { goto _test_eof27 }
	fallthrough
case 27:
	if data[p] == 110 { goto st28 }
	goto st0
st28:
	p++
	if p == pe { goto _test_eof28 }
	fallthrough
case 28:
	if data[p] == 32 { goto st29 }
	goto st0
st29:
	p++
	if p == pe { goto _test_eof29 }
	fallthrough
case 29:
	switch data[p] {
		case 32: goto st29
		case 61: goto st30
	}
	goto st0
st30:
	p++
	if p == pe { goto _test_eof30 }
	fallthrough
case 30:
	if data[p] == 32 { goto st31 }
	goto st0
st31:
	p++
	if p == pe { goto _test_eof31 }
	fallthrough
case 31:
	switch data[p] {
		case 32: goto st31
		case 102: goto tr57
		case 110: goto tr58
		case 111: goto tr59
		case 116: goto tr60
		case 121: goto tr61
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr56 }
	goto st0
tr56:
// line 41 "parser.rl"
	{ mark = p }
	goto st32
st32:
	p++
	if p == pe { goto _test_eof32 }
	fallthrough
case 32:
// line 650 "parser.go"
	switch data[p] {
		case 10: goto tr39
		case 32: goto tr62
	}
	goto st0
tr62:
// line 168 "parser.rl"
	{
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	goto st33
tr86:
// line 154 "parser.rl"
	{
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	goto st33
tr100:
// line 162 "parser.rl"
	{
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	goto st33
tr121:
// line 180 "parser.rl"
	{
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	goto st33
tr141:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st33
tr158:
// line 174 "parser.rl"
	{
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	goto st33
st33:
	p++
	if p == pe { goto _test_eof33 }
	fallthrough
case 33:
// line 698 "parser.go"
	switch data[p] {
		case 10: goto st25
		case 32: goto st33
	}
	goto st0
tr57:
// line 41 "parser.rl"
	{ mark = p }
	goto st34
st34:
	p++
	if p == pe { goto _test_eof34 }
	fallthrough
case 34:
// line 713 "parser.go"
	if data[p] == 97 { goto st35 }
	goto st0
st35:
	p++
	if p == pe { goto _test_eof35 }
	fallthrough
case 35:
	if data[p] == 108 { goto st36 }
	goto st0
st36:
	p++
	if p == pe { goto _test_eof36 }
	fallthrough
case 36:
	if data[p] == 115 { goto st37 }
	goto st0
st37:
	p++
	if p == pe { goto _test_eof37 }
	fallthrough
case 37:
	if data[p] == 101 { goto st32 }
	goto st0
tr58:
// line 41 "parser.rl"
	{ mark = p }
	goto st38
st38:
	p++
	if p == pe { goto _test_eof38 }
	fallthrough
case 38:
// line 746 "parser.go"
	if data[p] == 111 { goto st32 }
	goto st0
tr59:
// line 41 "parser.rl"
	{ mark = p }
	goto st39
st39:
	p++
	if p == pe { goto _test_eof39 }
	fallthrough
case 39:
// line 758 "parser.go"
	switch data[p] {
		case 102: goto st40
		case 110: goto st32
	}
	goto st0
st40:
	p++
	if p == pe { goto _test_eof40 }
	fallthrough
case 40:
	if data[p] == 102 { goto st32 }
	goto st0
tr60:
// line 41 "parser.rl"
	{ mark = p }
	goto st41
st41:
	p++
	if p == pe { goto _test_eof41 }
	fallthrough
case 41:
// line 780 "parser.go"
	if data[p] == 114 { goto st42 }
	goto st0
st42:
	p++
	if p == pe { goto _test_eof42 }
	fallthrough
case 42:
	if data[p] == 117 { goto st37 }
	goto st0
tr61:
// line 41 "parser.rl"
	{ mark = p }
	goto st43
st43:
	p++
	if p == pe { goto _test_eof43 }
	fallthrough
case 43:
// line 799 "parser.go"
	if data[p] == 101 { goto st44 }
	goto st0
st44:
	p++
	if p == pe { goto _test_eof44 }
	fallthrough
case 44:
	if data[p] == 115 { goto st32 }
	goto st0
st45:
	p++
	if p == pe { goto _test_eof45 }
	fallthrough
case 45:
	if data[p] == 111 { goto st46 }
	goto st0
st46:
	p++
	if p == pe { goto _test_eof46 }
	fallthrough
case 46:
	if data[p] == 115 { goto st47 }
	goto st0
st47:
	p++
	if p == pe { goto _test_eof47 }
	fallthrough
case 47:
	if data[p] == 116 { goto st48 }
	goto st0
st48:
	p++
	if p == pe { goto _test_eof48 }
	fallthrough
case 48:
	if data[p] == 109 { goto st49 }
	goto st0
st49:
	p++
	if p == pe { goto _test_eof49 }
	fallthrough
case 49:
	if data[p] == 97 { goto st50 }
	goto st0
st50:
	p++
	if p == pe { goto _test_eof50 }
	fallthrough
case 50:
	if data[p] == 115 { goto st51 }
	goto st0
st51:
	p++
	if p == pe { goto _test_eof51 }
	fallthrough
case 51:
	if data[p] == 107 { goto st52 }
	goto st0
st52:
	p++
	if p == pe { goto _test_eof52 }
	fallthrough
case 52:
	if data[p] == 32 { goto st53 }
	goto st0
st53:
	p++
	if p == pe { goto _test_eof53 }
	fallthrough
case 53:
	switch data[p] {
		case 32: goto st53
		case 61: goto st54
	}
	goto st0
st54:
	p++
	if p == pe { goto _test_eof54 }
	fallthrough
case 54:
	if data[p] == 32 { goto st55 }
	goto st0
st55:
	p++
	if p == pe { goto _test_eof55 }
	fallthrough
case 55:
	switch data[p] {
		case 32: goto st55
		case 42: goto tr81
		case 46: goto tr81
		case 63: goto tr81
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto tr81 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto tr81 }
	} else {
		goto tr81
	}
	goto st0
tr81:
// line 41 "parser.rl"
	{ mark = p }
	goto st56
st56:
	p++
	if p == pe { goto _test_eof56 }
	fallthrough
case 56:
// line 910 "parser.go"
	switch data[p] {
		case 42: goto st56
		case 46: goto st56
		case 64: goto st57
	}
	if data[p] < 63 {
		if 48 <= data[p] && data[p] <= 57 { goto st56 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st56 }
	} else {
		goto st56
	}
	goto st0
st57:
	p++
	if p == pe { goto _test_eof57 }
	fallthrough
case 57:
	switch data[p] {
		case 42: goto st58
		case 46: goto st58
		case 63: goto st58
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st58 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st58 }
	} else {
		goto st58
	}
	goto st0
st58:
	p++
	if p == pe { goto _test_eof58 }
	fallthrough
case 58:
	switch data[p] {
		case 10: goto tr85
		case 32: goto tr86
		case 42: goto st58
		case 46: goto st58
		case 63: goto st58
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st58 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st58 }
	} else {
		goto st58
	}
	goto st0
st59:
	p++
	if p == pe { goto _test_eof59 }
	fallthrough
case 59:
	if data[p] == 105 { goto st60 }
	goto st0
st60:
	p++
	if p == pe { goto _test_eof60 }
	fallthrough
case 60:
	if data[p] == 108 { goto st61 }
	goto st0
st61:
	p++
	if p == pe { goto _test_eof61 }
	fallthrough
case 61:
	if data[p] == 108 { goto st62 }
	goto st0
st62:
	p++
	if p == pe { goto _test_eof62 }
	fallthrough
case 62:
	if data[p] == 32 { goto st63 }
	goto st0
st63:
	p++
	if p == pe { goto _test_eof63 }
	fallthrough
case 63:
	switch data[p] {
		case 32: goto st63
		case 61: goto st64
	}
	goto st0
st64:
	p++
	if p == pe { goto _test_eof64 }
	fallthrough
case 64:
	if data[p] == 32 { goto st65 }
	goto st0
st65:
	p++
	if p == pe { goto _test_eof65 }
	fallthrough
case 65:
	switch data[p] {
		case 32: goto st65
		case 102: goto tr94
		case 110: goto tr95
		case 111: goto tr96
		case 116: goto tr97
		case 121: goto tr98
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr93 }
	goto st0
tr93:
// line 41 "parser.rl"
	{ mark = p }
	goto st66
st66:
	p++
	if p == pe { goto _test_eof66 }
	fallthrough
case 66:
// line 1031 "parser.go"
	switch data[p] {
		case 10: goto tr99
		case 32: goto tr100
	}
	goto st0
tr94:
// line 41 "parser.rl"
	{ mark = p }
	goto st67
st67:
	p++
	if p == pe { goto _test_eof67 }
	fallthrough
case 67:
// line 1046 "parser.go"
	if data[p] == 97 { goto st68 }
	goto st0
st68:
	p++
	if p == pe { goto _test_eof68 }
	fallthrough
case 68:
	if data[p] == 108 { goto st69 }
	goto st0
st69:
	p++
	if p == pe { goto _test_eof69 }
	fallthrough
case 69:
	if data[p] == 115 { goto st70 }
	goto st0
st70:
	p++
	if p == pe { goto _test_eof70 }
	fallthrough
case 70:
	if data[p] == 101 { goto st66 }
	goto st0
tr95:
// line 41 "parser.rl"
	{ mark = p }
	goto st71
st71:
	p++
	if p == pe { goto _test_eof71 }
	fallthrough
case 71:
// line 1079 "parser.go"
	if data[p] == 111 { goto st66 }
	goto st0
tr96:
// line 41 "parser.rl"
	{ mark = p }
	goto st72
st72:
	p++
	if p == pe { goto _test_eof72 }
	fallthrough
case 72:
// line 1091 "parser.go"
	switch data[p] {
		case 102: goto st73
		case 110: goto st66
	}
	goto st0
st73:
	p++
	if p == pe { goto _test_eof73 }
	fallthrough
case 73:
	if data[p] == 102 { goto st66 }
	goto st0
tr97:
// line 41 "parser.rl"
	{ mark = p }
	goto st74
st74:
	p++
	if p == pe { goto _test_eof74 }
	fallthrough
case 74:
// line 1113 "parser.go"
	if data[p] == 114 { goto st75 }
	goto st0
st75:
	p++
	if p == pe { goto _test_eof75 }
	fallthrough
case 75:
	if data[p] == 117 { goto st70 }
	goto st0
tr98:
// line 41 "parser.rl"
	{ mark = p }
	goto st76
st76:
	p++
	if p == pe { goto _test_eof76 }
	fallthrough
case 76:
// line 1132 "parser.go"
	if data[p] == 101 { goto st77 }
	goto st0
st77:
	p++
	if p == pe { goto _test_eof77 }
	fallthrough
case 77:
	if data[p] == 115 { goto st66 }
	goto st0
st78:
	p++
	if p == pe { goto _test_eof78 }
	fallthrough
case 78:
	if data[p] == 105 { goto st79 }
	goto st0
st79:
	p++
	if p == pe { goto _test_eof79 }
	fallthrough
case 79:
	if data[p] == 110 { goto st80 }
	goto st0
st80:
	p++
	if p == pe { goto _test_eof80 }
	fallthrough
case 80:
	if data[p] == 107 { goto st81 }
	goto st0
st81:
	p++
	if p == pe { goto _test_eof81 }
	fallthrough
case 81:
	if data[p] == 32 { goto st82 }
	goto st0
st82:
	p++
	if p == pe { goto _test_eof82 }
	fallthrough
case 82:
	switch data[p] {
		case 32: goto st82
		case 61: goto st83
	}
	goto st0
st83:
	p++
	if p == pe { goto _test_eof83 }
	fallthrough
case 83:
	if data[p] == 32 { goto st84 }
	goto st0
st84:
	p++
	if p == pe { goto _test_eof84 }
	fallthrough
case 84:
	switch data[p] {
		case 32: goto st84
		case 102: goto tr115
		case 110: goto tr116
		case 111: goto tr117
		case 116: goto tr118
		case 121: goto tr119
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr114 }
	goto st0
tr114:
// line 41 "parser.rl"
	{ mark = p }
	goto st85
st85:
	p++
	if p == pe { goto _test_eof85 }
	fallthrough
case 85:
// line 1211 "parser.go"
	switch data[p] {
		case 10: goto tr120
		case 32: goto tr121
	}
	goto st0
tr115:
// line 41 "parser.rl"
	{ mark = p }
	goto st86
st86:
	p++
	if p == pe { goto _test_eof86 }
	fallthrough
case 86:
// line 1226 "parser.go"
	if data[p] == 97 { goto st87 }
	goto st0
st87:
	p++
	if p == pe { goto _test_eof87 }
	fallthrough
case 87:
	if data[p] == 108 { goto st88 }
	goto st0
st88:
	p++
	if p == pe { goto _test_eof88 }
	fallthrough
case 88:
	if data[p] == 115 { goto st89 }
	goto st0
st89:
	p++
	if p == pe { goto _test_eof89 }
	fallthrough
case 89:
	if data[p] == 101 { goto st85 }
	goto st0
tr116:
// line 41 "parser.rl"
	{ mark = p }
	goto st90
st90:
	p++
	if p == pe { goto _test_eof90 }
	fallthrough
case 90:
// line 1259 "parser.go"
	if data[p] == 111 { goto st85 }
	goto st0
tr117:
// line 41 "parser.rl"
	{ mark = p }
	goto st91
st91:
	p++
	if p == pe { goto _test_eof91 }
	fallthrough
case 91:
// line 1271 "parser.go"
	switch data[p] {
		case 102: goto st92
		case 110: goto st85
	}
	goto st0
st92:
	p++
	if p == pe { goto _test_eof92 }
	fallthrough
case 92:
	if data[p] == 102 { goto st85 }
	goto st0
tr118:
// line 41 "parser.rl"
	{ mark = p }
	goto st93
st93:
	p++
	if p == pe { goto _test_eof93 }
	fallthrough
case 93:
// line 1293 "parser.go"
	if data[p] == 114 { goto st94 }
	goto st0
st94:
	p++
	if p == pe { goto _test_eof94 }
	fallthrough
case 94:
	if data[p] == 117 { goto st89 }
	goto st0
tr119:
// line 41 "parser.rl"
	{ mark = p }
	goto st95
st95:
	p++
	if p == pe { goto _test_eof95 }
	fallthrough
case 95:
// line 1312 "parser.go"
	if data[p] == 101 { goto st96 }
	goto st0
st96:
	p++
	if p == pe { goto _test_eof96 }
	fallthrough
case 96:
	if data[p] == 115 { goto st85 }
	goto st0
st97:
	p++
	if p == pe { goto _test_eof97 }
	fallthrough
case 97:
	if data[p] == 97 { goto st98 }
	goto st0
st98:
	p++
	if p == pe { goto _test_eof98 }
	fallthrough
case 98:
	if data[p] == 115 { goto st99 }
	goto st0
st99:
	p++
	if p == pe { goto _test_eof99 }
	fallthrough
case 99:
	if data[p] == 115 { goto st100 }
	goto st0
st100:
	p++
	if p == pe { goto _test_eof100 }
	fallthrough
case 100:
	if data[p] == 119 { goto st101 }
	goto st0
st101:
	p++
	if p == pe { goto _test_eof101 }
	fallthrough
case 101:
	if data[p] == 111 { goto st102 }
	goto st0
st102:
	p++
	if p == pe { goto _test_eof102 }
	fallthrough
case 102:
	if data[p] == 114 { goto st103 }
	goto st0
st103:
	p++
	if p == pe { goto _test_eof103 }
	fallthrough
case 103:
	if data[p] == 100 { goto st104 }
	goto st0
st104:
	p++
	if p == pe { goto _test_eof104 }
	fallthrough
case 104:
	if data[p] == 32 { goto st105 }
	goto st0
st105:
	p++
	if p == pe { goto _test_eof105 }
	fallthrough
case 105:
	switch data[p] {
		case 32: goto st105
		case 61: goto st106
	}
	goto st0
st106:
	p++
	if p == pe { goto _test_eof106 }
	fallthrough
case 106:
	if data[p] == 32 { goto st107 }
	goto st0
st107:
	p++
	if p == pe { goto _test_eof107 }
	fallthrough
case 107:
	if data[p] == 32 { goto st107 }
	if 33 <= data[p] && data[p] <= 126 { goto tr139 }
	goto st0
tr139:
// line 41 "parser.rl"
	{ mark = p }
	goto st108
st108:
	p++
	if p == pe { goto _test_eof108 }
	fallthrough
case 108:
// line 1412 "parser.go"
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr141
	}
	if 33 <= data[p] && data[p] <= 126 { goto st108 }
	goto st0
st109:
	p++
	if p == pe { goto _test_eof109 }
	fallthrough
case 109:
	if data[p] == 101 { goto st110 }
	goto st0
st110:
	p++
	if p == pe { goto _test_eof110 }
	fallthrough
case 110:
	if data[p] == 110 { goto st111 }
	goto st0
st111:
	p++
	if p == pe { goto _test_eof111 }
	fallthrough
case 111:
	if data[p] == 105 { goto st112 }
	goto st0
st112:
	p++
	if p == pe { goto _test_eof112 }
	fallthrough
case 112:
	if data[p] == 99 { goto st113 }
	goto st0
st113:
	p++
	if p == pe { goto _test_eof113 }
	fallthrough
case 113:
	if data[p] == 107 { goto st114 }
	goto st0
st114:
	p++
	if p == pe { goto _test_eof114 }
	fallthrough
case 114:
	if data[p] == 32 { goto st115 }
	goto st0
st115:
	p++
	if p == pe { goto _test_eof115 }
	fallthrough
case 115:
	switch data[p] {
		case 32: goto st115
		case 61: goto st116
	}
	goto st0
st116:
	p++
	if p == pe { goto _test_eof116 }
	fallthrough
case 116:
	if data[p] == 32 { goto st117 }
	goto st0
st117:
	p++
	if p == pe { goto _test_eof117 }
	fallthrough
case 117:
	switch data[p] {
		case 32: goto st117
		case 102: goto tr152
		case 110: goto tr153
		case 111: goto tr154
		case 116: goto tr155
		case 121: goto tr156
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr151 }
	goto st0
tr151:
// line 41 "parser.rl"
	{ mark = p }
	goto st118
st118:
	p++
	if p == pe { goto _test_eof118 }
	fallthrough
case 118:
// line 1502 "parser.go"
	switch data[p] {
		case 10: goto tr157
		case 32: goto tr158
	}
	goto st0
tr152:
// line 41 "parser.rl"
	{ mark = p }
	goto st119
st119:
	p++
	if p == pe { goto _test_eof119 }
	fallthrough
case 119:
// line 1517 "parser.go"
	if data[p] == 97 { goto st120 }
	goto st0
st120:
	p++
	if p == pe { goto _test_eof120 }
	fallthrough
case 120:
	if data[p] == 108 { goto st121 }
	goto st0
st121:
	p++
	if p == pe { goto _test_eof121 }
	fallthrough
case 121:
	if data[p] == 115 { goto st122 }
	goto st0
st122:
	p++
	if p == pe { goto _test_eof122 }
	fallthrough
case 122:
	if data[p] == 101 { goto st118 }
	goto st0
tr153:
// line 41 "parser.rl"
	{ mark = p }
	goto st123
st123:
	p++
	if p == pe { goto _test_eof123 }
	fallthrough
case 123:
// line 1550 "parser.go"
	if data[p] == 111 { goto st118 }
	goto st0
tr154:
// line 41 "parser.rl"
	{ mark = p }
	goto st124
st124:
	p++
	if p == pe { goto _test_eof124 }
	fallthrough
case 124:
// line 1562 "parser.go"
	switch data[p] {
		case 102: goto st125
		case 110: goto st118
	}
	goto st0
st125:
	p++
	if p == pe { goto _test_eof125 }
	fallthrough
case 125:
	if data[p] == 102 { goto st118 }
	goto st0
tr155:
// line 41 "parser.rl"
	{ mark = p }
	goto st126
st126:
	p++
	if p == pe { goto _test_eof126 }
	fallthrough
case 126:
// line 1584 "parser.go"
	if data[p] == 114 { goto st127 }
	goto st0
st127:
	p++
	if p == pe { goto _test_eof127 }
	fallthrough
case 127:
	if data[p] == 117 { goto st122 }
	goto st0
tr156:
// line 41 "parser.rl"
	{ mark = p }
	goto st128
st128:
	p++
	if p == pe { goto _test_eof128 }
	fallthrough
case 128:
// line 1603 "parser.go"
	if data[p] == 101 { goto st129 }
	goto st0
st129:
	p++
	if p == pe { goto _test_eof129 }
	fallthrough
case 129:
	if data[p] == 115 { goto st118 }
	goto st0
tr40:
// line 168 "parser.rl"
	{
		cur.(*cOper).CanBan = getbool(data[mark:p])
	}
	goto st130
tr189:
// line 154 "parser.rl"
	{
		cur.(*cOper).HostMask = append(
			cur.(*cOper).HostMask, string(data[mark:p]))
	}
	goto st130
tr204:
// line 162 "parser.rl"
	{
		cur.(*cOper).CanKill = getbool(data[mark:p])
	}
	goto st130
tr226:
// line 180 "parser.rl"
	{
		cur.(*cOper).CanLink = getbool(data[mark:p])
	}
	goto st130
tr247:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st130
tr275:
// line 174 "parser.rl"
	{
		cur.(*cOper).CanRenick = getbool(data[mark:p])
	}
	goto st130
st130:
	p++
	if p == pe { goto _test_eof130 }
	fallthrough
case 130:
// line 1655 "parser.go"
	switch data[p] {
		case 10: goto st25
		case 32: goto st130
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st24 }
	goto st0
tr33:
// line 41 "parser.rl"
	{ mark = p }
	goto st131
st131:
	p++
	if p == pe { goto _test_eof131 }
	fallthrough
case 131:
// line 1672 "parser.go"
	if data[p] == 97 { goto st132 }
	goto st0
st132:
	p++
	if p == pe { goto _test_eof132 }
	fallthrough
case 132:
	if data[p] == 108 { goto st133 }
	goto st0
st133:
	p++
	if p == pe { goto _test_eof133 }
	fallthrough
case 133:
	if data[p] == 115 { goto st134 }
	goto st0
st134:
	p++
	if p == pe { goto _test_eof134 }
	fallthrough
case 134:
	if data[p] == 101 { goto st23 }
	goto st0
tr34:
// line 41 "parser.rl"
	{ mark = p }
	goto st135
st135:
	p++
	if p == pe { goto _test_eof135 }
	fallthrough
case 135:
// line 1705 "parser.go"
	if data[p] == 111 { goto st23 }
	goto st0
tr35:
// line 41 "parser.rl"
	{ mark = p }
	goto st136
st136:
	p++
	if p == pe { goto _test_eof136 }
	fallthrough
case 136:
// line 1717 "parser.go"
	switch data[p] {
		case 102: goto st137
		case 110: goto st23
	}
	goto st0
st137:
	p++
	if p == pe { goto _test_eof137 }
	fallthrough
case 137:
	if data[p] == 102 { goto st23 }
	goto st0
tr36:
// line 41 "parser.rl"
	{ mark = p }
	goto st138
st138:
	p++
	if p == pe { goto _test_eof138 }
	fallthrough
case 138:
// line 1739 "parser.go"
	if data[p] == 114 { goto st139 }
	goto st0
st139:
	p++
	if p == pe { goto _test_eof139 }
	fallthrough
case 139:
	if data[p] == 117 { goto st134 }
	goto st0
tr37:
// line 41 "parser.rl"
	{ mark = p }
	goto st140
st140:
	p++
	if p == pe { goto _test_eof140 }
	fallthrough
case 140:
// line 1758 "parser.go"
	if data[p] == 101 { goto st141 }
	goto st0
st141:
	p++
	if p == pe { goto _test_eof141 }
	fallthrough
case 141:
	if data[p] == 115 { goto st23 }
	goto st0
st142:
	p++
	if p == pe { goto _test_eof142 }
	fallthrough
case 142:
	if data[p] == 111 { goto st143 }
	goto st0
st143:
	p++
	if p == pe { goto _test_eof143 }
	fallthrough
case 143:
	if data[p] == 115 { goto st144 }
	goto st0
st144:
	p++
	if p == pe { goto _test_eof144 }
	fallthrough
case 144:
	if data[p] == 116 { goto st145 }
	goto st0
st145:
	p++
	if p == pe { goto _test_eof145 }
	fallthrough
case 145:
	if data[p] == 109 { goto st146 }
	goto st0
st146:
	p++
	if p == pe { goto _test_eof146 }
	fallthrough
case 146:
	if data[p] == 97 { goto st147 }
	goto st0
st147:
	p++
	if p == pe { goto _test_eof147 }
	fallthrough
case 147:
	if data[p] == 115 { goto st148 }
	goto st0
st148:
	p++
	if p == pe { goto _test_eof148 }
	fallthrough
case 148:
	if data[p] == 107 { goto st149 }
	goto st0
st149:
	p++
	if p == pe { goto _test_eof149 }
	fallthrough
case 149:
	if data[p] == 32 { goto st150 }
	goto st0
st150:
	p++
	if p == pe { goto _test_eof150 }
	fallthrough
case 150:
	switch data[p] {
		case 32: goto st150
		case 61: goto st151
	}
	goto st0
st151:
	p++
	if p == pe { goto _test_eof151 }
	fallthrough
case 151:
	if data[p] == 32 { goto st152 }
	goto st0
st152:
	p++
	if p == pe { goto _test_eof152 }
	fallthrough
case 152:
	switch data[p] {
		case 32: goto st152
		case 42: goto tr184
		case 46: goto tr184
		case 63: goto tr184
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto tr184 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto tr184 }
	} else {
		goto tr184
	}
	goto st0
tr184:
// line 41 "parser.rl"
	{ mark = p }
	goto st153
st153:
	p++
	if p == pe { goto _test_eof153 }
	fallthrough
case 153:
// line 1869 "parser.go"
	switch data[p] {
		case 42: goto st153
		case 46: goto st153
		case 64: goto st154
	}
	if data[p] < 63 {
		if 48 <= data[p] && data[p] <= 57 { goto st153 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st153 }
	} else {
		goto st153
	}
	goto st0
st154:
	p++
	if p == pe { goto _test_eof154 }
	fallthrough
case 154:
	switch data[p] {
		case 42: goto st155
		case 46: goto st155
		case 63: goto st155
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st155 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto st155 }
	} else {
		goto st155
	}
	goto st0
st155:
	p++
	if p == pe { goto _test_eof155 }
	fallthrough
case 155:
	switch data[p] {
		case 10: goto tr85
		case 32: goto tr189
		case 42: goto st155
		case 46: goto st155
		case 63: goto st155
		case 125: goto tr190
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr188 }
	} else if data[p] > 57 {
		if data[p] > 90 {
			if 97 <= data[p] && data[p] <= 122 { goto st155 }
		} else if data[p] >= 65 {
			goto st155
		}
	} else {
		goto st155
	}
	goto st0
st156:
	p++
	if p == pe { goto _test_eof156 }
	fallthrough
case 156:
	if data[p] == 105 { goto st157 }
	goto st0
st157:
	p++
	if p == pe { goto _test_eof157 }
	fallthrough
case 157:
	if data[p] == 108 { goto st158 }
	goto st0
st158:
	p++
	if p == pe { goto _test_eof158 }
	fallthrough
case 158:
	if data[p] == 108 { goto st159 }
	goto st0
st159:
	p++
	if p == pe { goto _test_eof159 }
	fallthrough
case 159:
	if data[p] == 32 { goto st160 }
	goto st0
st160:
	p++
	if p == pe { goto _test_eof160 }
	fallthrough
case 160:
	switch data[p] {
		case 32: goto st160
		case 61: goto st161
	}
	goto st0
st161:
	p++
	if p == pe { goto _test_eof161 }
	fallthrough
case 161:
	if data[p] == 32 { goto st162 }
	goto st0
st162:
	p++
	if p == pe { goto _test_eof162 }
	fallthrough
case 162:
	switch data[p] {
		case 32: goto st162
		case 102: goto tr198
		case 110: goto tr199
		case 111: goto tr200
		case 116: goto tr201
		case 121: goto tr202
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr197 }
	goto st0
tr197:
// line 41 "parser.rl"
	{ mark = p }
	goto st163
st163:
	p++
	if p == pe { goto _test_eof163 }
	fallthrough
case 163:
// line 1995 "parser.go"
	switch data[p] {
		case 10: goto tr99
		case 32: goto tr204
		case 125: goto tr205
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr203 }
	goto st0
tr198:
// line 41 "parser.rl"
	{ mark = p }
	goto st164
st164:
	p++
	if p == pe { goto _test_eof164 }
	fallthrough
case 164:
// line 2012 "parser.go"
	if data[p] == 97 { goto st165 }
	goto st0
st165:
	p++
	if p == pe { goto _test_eof165 }
	fallthrough
case 165:
	if data[p] == 108 { goto st166 }
	goto st0
st166:
	p++
	if p == pe { goto _test_eof166 }
	fallthrough
case 166:
	if data[p] == 115 { goto st167 }
	goto st0
st167:
	p++
	if p == pe { goto _test_eof167 }
	fallthrough
case 167:
	if data[p] == 101 { goto st163 }
	goto st0
tr199:
// line 41 "parser.rl"
	{ mark = p }
	goto st168
st168:
	p++
	if p == pe { goto _test_eof168 }
	fallthrough
case 168:
// line 2045 "parser.go"
	if data[p] == 111 { goto st163 }
	goto st0
tr200:
// line 41 "parser.rl"
	{ mark = p }
	goto st169
st169:
	p++
	if p == pe { goto _test_eof169 }
	fallthrough
case 169:
// line 2057 "parser.go"
	switch data[p] {
		case 102: goto st170
		case 110: goto st163
	}
	goto st0
st170:
	p++
	if p == pe { goto _test_eof170 }
	fallthrough
case 170:
	if data[p] == 102 { goto st163 }
	goto st0
tr201:
// line 41 "parser.rl"
	{ mark = p }
	goto st171
st171:
	p++
	if p == pe { goto _test_eof171 }
	fallthrough
case 171:
// line 2079 "parser.go"
	if data[p] == 114 { goto st172 }
	goto st0
st172:
	p++
	if p == pe { goto _test_eof172 }
	fallthrough
case 172:
	if data[p] == 117 { goto st167 }
	goto st0
tr202:
// line 41 "parser.rl"
	{ mark = p }
	goto st173
st173:
	p++
	if p == pe { goto _test_eof173 }
	fallthrough
case 173:
// line 2098 "parser.go"
	if data[p] == 101 { goto st174 }
	goto st0
st174:
	p++
	if p == pe { goto _test_eof174 }
	fallthrough
case 174:
	if data[p] == 115 { goto st163 }
	goto st0
st175:
	p++
	if p == pe { goto _test_eof175 }
	fallthrough
case 175:
	if data[p] == 105 { goto st176 }
	goto st0
st176:
	p++
	if p == pe { goto _test_eof176 }
	fallthrough
case 176:
	if data[p] == 110 { goto st177 }
	goto st0
st177:
	p++
	if p == pe { goto _test_eof177 }
	fallthrough
case 177:
	if data[p] == 107 { goto st178 }
	goto st0
st178:
	p++
	if p == pe { goto _test_eof178 }
	fallthrough
case 178:
	if data[p] == 32 { goto st179 }
	goto st0
st179:
	p++
	if p == pe { goto _test_eof179 }
	fallthrough
case 179:
	switch data[p] {
		case 32: goto st179
		case 61: goto st180
	}
	goto st0
st180:
	p++
	if p == pe { goto _test_eof180 }
	fallthrough
case 180:
	if data[p] == 32 { goto st181 }
	goto st0
st181:
	p++
	if p == pe { goto _test_eof181 }
	fallthrough
case 181:
	switch data[p] {
		case 32: goto st181
		case 102: goto tr220
		case 110: goto tr221
		case 111: goto tr222
		case 116: goto tr223
		case 121: goto tr224
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr219 }
	goto st0
tr219:
// line 41 "parser.rl"
	{ mark = p }
	goto st182
st182:
	p++
	if p == pe { goto _test_eof182 }
	fallthrough
case 182:
// line 2177 "parser.go"
	switch data[p] {
		case 10: goto tr120
		case 32: goto tr226
		case 125: goto tr227
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr225 }
	goto st0
tr220:
// line 41 "parser.rl"
	{ mark = p }
	goto st183
st183:
	p++
	if p == pe { goto _test_eof183 }
	fallthrough
case 183:
// line 2194 "parser.go"
	if data[p] == 97 { goto st184 }
	goto st0
st184:
	p++
	if p == pe { goto _test_eof184 }
	fallthrough
case 184:
	if data[p] == 108 { goto st185 }
	goto st0
st185:
	p++
	if p == pe { goto _test_eof185 }
	fallthrough
case 185:
	if data[p] == 115 { goto st186 }
	goto st0
st186:
	p++
	if p == pe { goto _test_eof186 }
	fallthrough
case 186:
	if data[p] == 101 { goto st182 }
	goto st0
tr221:
// line 41 "parser.rl"
	{ mark = p }
	goto st187
st187:
	p++
	if p == pe { goto _test_eof187 }
	fallthrough
case 187:
// line 2227 "parser.go"
	if data[p] == 111 { goto st182 }
	goto st0
tr222:
// line 41 "parser.rl"
	{ mark = p }
	goto st188
st188:
	p++
	if p == pe { goto _test_eof188 }
	fallthrough
case 188:
// line 2239 "parser.go"
	switch data[p] {
		case 102: goto st189
		case 110: goto st182
	}
	goto st0
st189:
	p++
	if p == pe { goto _test_eof189 }
	fallthrough
case 189:
	if data[p] == 102 { goto st182 }
	goto st0
tr223:
// line 41 "parser.rl"
	{ mark = p }
	goto st190
st190:
	p++
	if p == pe { goto _test_eof190 }
	fallthrough
case 190:
// line 2261 "parser.go"
	if data[p] == 114 { goto st191 }
	goto st0
st191:
	p++
	if p == pe { goto _test_eof191 }
	fallthrough
case 191:
	if data[p] == 117 { goto st186 }
	goto st0
tr224:
// line 41 "parser.rl"
	{ mark = p }
	goto st192
st192:
	p++
	if p == pe { goto _test_eof192 }
	fallthrough
case 192:
// line 2280 "parser.go"
	if data[p] == 101 { goto st193 }
	goto st0
st193:
	p++
	if p == pe { goto _test_eof193 }
	fallthrough
case 193:
	if data[p] == 115 { goto st182 }
	goto st0
st194:
	p++
	if p == pe { goto _test_eof194 }
	fallthrough
case 194:
	if data[p] == 97 { goto st195 }
	goto st0
st195:
	p++
	if p == pe { goto _test_eof195 }
	fallthrough
case 195:
	if data[p] == 115 { goto st196 }
	goto st0
st196:
	p++
	if p == pe { goto _test_eof196 }
	fallthrough
case 196:
	if data[p] == 115 { goto st197 }
	goto st0
st197:
	p++
	if p == pe { goto _test_eof197 }
	fallthrough
case 197:
	if data[p] == 119 { goto st198 }
	goto st0
st198:
	p++
	if p == pe { goto _test_eof198 }
	fallthrough
case 198:
	if data[p] == 111 { goto st199 }
	goto st0
st199:
	p++
	if p == pe { goto _test_eof199 }
	fallthrough
case 199:
	if data[p] == 114 { goto st200 }
	goto st0
st200:
	p++
	if p == pe { goto _test_eof200 }
	fallthrough
case 200:
	if data[p] == 100 { goto st201 }
	goto st0
st201:
	p++
	if p == pe { goto _test_eof201 }
	fallthrough
case 201:
	if data[p] == 32 { goto st202 }
	goto st0
st202:
	p++
	if p == pe { goto _test_eof202 }
	fallthrough
case 202:
	switch data[p] {
		case 32: goto st202
		case 61: goto st203
	}
	goto st0
st203:
	p++
	if p == pe { goto _test_eof203 }
	fallthrough
case 203:
	if data[p] == 32 { goto st204 }
	goto st0
st204:
	p++
	if p == pe { goto _test_eof204 }
	fallthrough
case 204:
	if data[p] == 32 { goto st204 }
	if 33 <= data[p] && data[p] <= 126 { goto tr245 }
	goto st0
tr245:
// line 41 "parser.rl"
	{ mark = p }
	goto st205
st205:
	p++
	if p == pe { goto _test_eof205 }
	fallthrough
case 205:
// line 2380 "parser.go"
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
tr249:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st563
st563:
	p++
	if p == pe { goto _test_eof563 }
	fallthrough
case 563:
// line 2403 "parser.go"
	switch data[p] {
		case 10: goto tr639
		case 32: goto tr640
		case 111: goto tr641
		case 112: goto tr642
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr638
	}
	goto st0
tr638:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
	goto st564
st564:
	p++
	if p == pe { goto _test_eof564 }
	fallthrough
case 564:
// line 2434 "parser.go"
	switch data[p] {
		case 32: goto st564
		case 111: goto tr2
		case 112: goto tr3
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st564 }
	goto st0
tr639:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
	goto st565
st565:
	p++
	if p == pe { goto _test_eof565 }
	fallthrough
case 565:
// line 2460 "parser.go"
	switch data[p] {
		case 32: goto st565
		case 98: goto st26
		case 104: goto st45
		case 107: goto st59
		case 108: goto st78
		case 111: goto tr2
		case 112: goto tr645
		case 114: goto st109
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st565 }
	goto st0
tr645:
// line 77 "parser.rl"
	{ cur = defaultPort() }
	goto st206
st206:
	p++
	if p == pe { goto _test_eof206 }
	fallthrough
case 206:
// line 2483 "parser.go"
	switch data[p] {
		case 97: goto st98
		case 111: goto st4
	}
	goto st0
tr640:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
	goto st566
st566:
	p++
	if p == pe { goto _test_eof566 }
	fallthrough
case 566:
// line 2506 "parser.go"
	switch data[p] {
		case 10: goto st565
		case 32: goto st566
		case 111: goto tr2
		case 112: goto tr3
		case 125: goto st562
	}
	if 9 <= data[p] && data[p] <= 13 { goto st564 }
	goto st0
tr641:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
// line 133 "parser.rl"
	{ cur = defaultOper() }
	goto st207
st207:
	p++
	if p == pe { goto _test_eof207 }
	fallthrough
case 207:
// line 2531 "parser.go"
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 112: goto st208
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st208:
	p++
	if p == pe { goto _test_eof208 }
	fallthrough
case 208:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 101: goto st209
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st209:
	p++
	if p == pe { goto _test_eof209 }
	fallthrough
case 209:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 114: goto st210
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st210:
	p++
	if p == pe { goto _test_eof210 }
	fallthrough
case 210:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr253
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
tr253:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st211
st211:
	p++
	if p == pe { goto _test_eof211 }
	fallthrough
case 211:
// line 2605 "parser.go"
	switch data[p] {
		case 10: goto st25
		case 32: goto st211
		case 125: goto st562
	}
	if data[p] < 65 {
		if 9 <= data[p] && data[p] <= 13 { goto st24 }
	} else if data[p] > 90 {
		if 97 <= data[p] && data[p] <= 122 { goto tr16 }
	} else {
		goto tr16
	}
	goto st0
tr642:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
// line 77 "parser.rl"
	{ cur = defaultPort() }
	goto st212
st212:
	p++
	if p == pe { goto _test_eof212 }
	fallthrough
case 212:
// line 2634 "parser.go"
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 111: goto st213
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st213:
	p++
	if p == pe { goto _test_eof213 }
	fallthrough
case 213:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 114: goto st214
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st214:
	p++
	if p == pe { goto _test_eof214 }
	fallthrough
case 214:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr247
		case 116: goto st215
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
st215:
	p++
	if p == pe { goto _test_eof215 }
	fallthrough
case 215:
	switch data[p] {
		case 10: goto tr140
		case 32: goto tr258
		case 125: goto tr249
	}
	if data[p] > 13 {
		if 33 <= data[p] && data[p] <= 126 { goto st205 }
	} else if data[p] >= 9 {
		goto tr246
	}
	goto st0
tr258:
// line 147 "parser.rl"
	{
		cur.(*cOper).Password = string(data[mark:p])
	}
	goto st216
st216:
	p++
	if p == pe { goto _test_eof216 }
	fallthrough
case 216:
// line 2708 "parser.go"
	switch data[p] {
		case 10: goto st25
		case 32: goto st216
		case 125: goto st562
	}
	if data[p] > 13 {
		if 48 <= data[p] && data[p] <= 57 { goto tr8 }
	} else if data[p] >= 9 {
		goto st24
	}
	goto st0
st217:
	p++
	if p == pe { goto _test_eof217 }
	fallthrough
case 217:
	if data[p] == 101 { goto st218 }
	goto st0
st218:
	p++
	if p == pe { goto _test_eof218 }
	fallthrough
case 218:
	if data[p] == 110 { goto st219 }
	goto st0
st219:
	p++
	if p == pe { goto _test_eof219 }
	fallthrough
case 219:
	if data[p] == 105 { goto st220 }
	goto st0
st220:
	p++
	if p == pe { goto _test_eof220 }
	fallthrough
case 220:
	if data[p] == 99 { goto st221 }
	goto st0
st221:
	p++
	if p == pe { goto _test_eof221 }
	fallthrough
case 221:
	if data[p] == 107 { goto st222 }
	goto st0
st222:
	p++
	if p == pe { goto _test_eof222 }
	fallthrough
case 222:
	if data[p] == 32 { goto st223 }
	goto st0
st223:
	p++
	if p == pe { goto _test_eof223 }
	fallthrough
case 223:
	switch data[p] {
		case 32: goto st223
		case 61: goto st224
	}
	goto st0
st224:
	p++
	if p == pe { goto _test_eof224 }
	fallthrough
case 224:
	if data[p] == 32 { goto st225 }
	goto st0
st225:
	p++
	if p == pe { goto _test_eof225 }
	fallthrough
case 225:
	switch data[p] {
		case 32: goto st225
		case 102: goto tr269
		case 110: goto tr270
		case 111: goto tr271
		case 116: goto tr272
		case 121: goto tr273
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr268 }
	goto st0
tr268:
// line 41 "parser.rl"
	{ mark = p }
	goto st226
st226:
	p++
	if p == pe { goto _test_eof226 }
	fallthrough
case 226:
// line 2803 "parser.go"
	switch data[p] {
		case 10: goto tr157
		case 32: goto tr275
		case 125: goto tr276
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr274 }
	goto st0
tr269:
// line 41 "parser.rl"
	{ mark = p }
	goto st227
st227:
	p++
	if p == pe { goto _test_eof227 }
	fallthrough
case 227:
// line 2820 "parser.go"
	if data[p] == 97 { goto st228 }
	goto st0
st228:
	p++
	if p == pe { goto _test_eof228 }
	fallthrough
case 228:
	if data[p] == 108 { goto st229 }
	goto st0
st229:
	p++
	if p == pe { goto _test_eof229 }
	fallthrough
case 229:
	if data[p] == 115 { goto st230 }
	goto st0
st230:
	p++
	if p == pe { goto _test_eof230 }
	fallthrough
case 230:
	if data[p] == 101 { goto st226 }
	goto st0
tr270:
// line 41 "parser.rl"
	{ mark = p }
	goto st231
st231:
	p++
	if p == pe { goto _test_eof231 }
	fallthrough
case 231:
// line 2853 "parser.go"
	if data[p] == 111 { goto st226 }
	goto st0
tr271:
// line 41 "parser.rl"
	{ mark = p }
	goto st232
st232:
	p++
	if p == pe { goto _test_eof232 }
	fallthrough
case 232:
// line 2865 "parser.go"
	switch data[p] {
		case 102: goto st233
		case 110: goto st226
	}
	goto st0
st233:
	p++
	if p == pe { goto _test_eof233 }
	fallthrough
case 233:
	if data[p] == 102 { goto st226 }
	goto st0
tr272:
// line 41 "parser.rl"
	{ mark = p }
	goto st234
st234:
	p++
	if p == pe { goto _test_eof234 }
	fallthrough
case 234:
// line 2887 "parser.go"
	if data[p] == 114 { goto st235 }
	goto st0
st235:
	p++
	if p == pe { goto _test_eof235 }
	fallthrough
case 235:
	if data[p] == 117 { goto st230 }
	goto st0
tr273:
// line 41 "parser.rl"
	{ mark = p }
	goto st236
st236:
	p++
	if p == pe { goto _test_eof236 }
	fallthrough
case 236:
// line 2906 "parser.go"
	if data[p] == 101 { goto st237 }
	goto st0
st237:
	p++
	if p == pe { goto _test_eof237 }
	fallthrough
case 237:
	if data[p] == 115 { goto st226 }
	goto st0
tr10:
// line 85 "parser.rl"
	{
		cur.(*cPort).Port = getint(data[mark:p])
	}
	goto st238
st238:
	p++
	if p == pe { goto _test_eof238 }
	fallthrough
case 238:
// line 2927 "parser.go"
	switch data[p] {
		case 10: goto st560
		case 32: goto st238
		case 123: goto st239
	}
	goto st0
st239:
	p++
	if p == pe { goto _test_eof239 }
	fallthrough
case 239:
	switch data[p] {
		case 32: goto st239
		case 98: goto st240
		case 99: goto st505
		case 115: goto st524
		case 122: goto st542
	}
	if 9 <= data[p] && data[p] <= 13 { goto st239 }
	goto st0
st240:
	p++
	if p == pe { goto _test_eof240 }
	fallthrough
case 240:
	if data[p] == 105 { goto st241 }
	goto st0
st241:
	p++
	if p == pe { goto _test_eof241 }
	fallthrough
case 241:
	if data[p] == 110 { goto st242 }
	goto st0
st242:
	p++
	if p == pe { goto _test_eof242 }
	fallthrough
case 242:
	if data[p] == 100 { goto st243 }
	goto st0
st243:
	p++
	if p == pe { goto _test_eof243 }
	fallthrough
case 243:
	if data[p] == 95 { goto st244 }
	goto st0
st244:
	p++
	if p == pe { goto _test_eof244 }
	fallthrough
case 244:
	if data[p] == 105 { goto st245 }
	goto st0
st245:
	p++
	if p == pe { goto _test_eof245 }
	fallthrough
case 245:
	if data[p] == 112 { goto st246 }
	goto st0
st246:
	p++
	if p == pe { goto _test_eof246 }
	fallthrough
case 246:
	if data[p] == 32 { goto st247 }
	goto st0
st247:
	p++
	if p == pe { goto _test_eof247 }
	fallthrough
case 247:
	switch data[p] {
		case 32: goto st247
		case 61: goto st248
	}
	goto st0
st248:
	p++
	if p == pe { goto _test_eof248 }
	fallthrough
case 248:
	if data[p] == 32 { goto st249 }
	goto st0
st249:
	p++
	if p == pe { goto _test_eof249 }
	fallthrough
case 249:
	switch data[p] {
		case 32: goto st249
		case 50: goto tr301
		case 58: goto tr303
	}
	if data[p] < 51 {
		if 48 <= data[p] && data[p] <= 49 { goto tr300 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto tr304 }
		} else if data[p] >= 65 {
			goto tr304
		}
	} else {
		goto tr302
	}
	goto st0
tr300:
// line 41 "parser.rl"
	{ mark = p }
	goto st250
st250:
	p++
	if p == pe { goto _test_eof250 }
	fallthrough
case 250:
// line 3045 "parser.go"
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st431 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st498 }
	} else {
		goto st498
	}
	goto st0
st251:
	p++
	if p == pe { goto _test_eof251 }
	fallthrough
case 251:
	if data[p] == 50 { goto st429 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st427 }
	} else if data[p] >= 48 {
		goto st252
	}
	goto st0
st252:
	p++
	if p == pe { goto _test_eof252 }
	fallthrough
case 252:
	if data[p] == 46 { goto st253 }
	if 48 <= data[p] && data[p] <= 57 { goto st427 }
	goto st0
st253:
	p++
	if p == pe { goto _test_eof253 }
	fallthrough
case 253:
	if data[p] == 50 { goto st425 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st423 }
	} else if data[p] >= 48 {
		goto st254
	}
	goto st0
st254:
	p++
	if p == pe { goto _test_eof254 }
	fallthrough
case 254:
	if data[p] == 46 { goto st255 }
	if 48 <= data[p] && data[p] <= 57 { goto st423 }
	goto st0
st255:
	p++
	if p == pe { goto _test_eof255 }
	fallthrough
case 255:
	if data[p] == 50 { goto st421 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st419 }
	} else if data[p] >= 48 {
		goto st256
	}
	goto st0
st256:
	p++
	if p == pe { goto _test_eof256 }
	fallthrough
case 256:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] > 13 {
		if 48 <= data[p] && data[p] <= 57 { goto st419 }
	} else if data[p] >= 9 {
		goto tr320
	}
	goto st0
tr320:
// line 91 "parser.rl"
	{
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	goto st257
tr582:
// line 98 "parser.rl"
	{
		cur.(*cPort).Class = string(data[mark:p])
	}
	goto st257
tr600:
// line 105 "parser.rl"
	{
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	goto st257
tr621:
// line 111 "parser.rl"
	{
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	goto st257
st257:
	p++
	if p == pe { goto _test_eof257 }
	fallthrough
case 257:
// line 3155 "parser.go"
	switch data[p] {
		case 32: goto st257
		case 125: goto st560
	}
	if 9 <= data[p] && data[p] <= 13 { goto st257 }
	goto st0
tr321:
// line 91 "parser.rl"
	{
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	goto st258
tr448:
// line 98 "parser.rl"
	{
		cur.(*cPort).Class = string(data[mark:p])
	}
	goto st258
tr465:
// line 105 "parser.rl"
	{
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	goto st258
tr485:
// line 111 "parser.rl"
	{
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	goto st258
st258:
	p++
	if p == pe { goto _test_eof258 }
	fallthrough
case 258:
// line 3191 "parser.go"
	switch data[p] {
		case 32: goto st258
		case 98: goto st259
		case 99: goto st363
		case 115: goto st382
		case 122: goto st400
		case 125: goto st560
	}
	if 9 <= data[p] && data[p] <= 13 { goto st258 }
	goto st0
st259:
	p++
	if p == pe { goto _test_eof259 }
	fallthrough
case 259:
	if data[p] == 105 { goto st260 }
	goto st0
st260:
	p++
	if p == pe { goto _test_eof260 }
	fallthrough
case 260:
	if data[p] == 110 { goto st261 }
	goto st0
st261:
	p++
	if p == pe { goto _test_eof261 }
	fallthrough
case 261:
	if data[p] == 100 { goto st262 }
	goto st0
st262:
	p++
	if p == pe { goto _test_eof262 }
	fallthrough
case 262:
	if data[p] == 95 { goto st263 }
	goto st0
st263:
	p++
	if p == pe { goto _test_eof263 }
	fallthrough
case 263:
	if data[p] == 105 { goto st264 }
	goto st0
st264:
	p++
	if p == pe { goto _test_eof264 }
	fallthrough
case 264:
	if data[p] == 112 { goto st265 }
	goto st0
st265:
	p++
	if p == pe { goto _test_eof265 }
	fallthrough
case 265:
	if data[p] == 32 { goto st266 }
	goto st0
st266:
	p++
	if p == pe { goto _test_eof266 }
	fallthrough
case 266:
	switch data[p] {
		case 32: goto st266
		case 61: goto st267
	}
	goto st0
st267:
	p++
	if p == pe { goto _test_eof267 }
	fallthrough
case 267:
	if data[p] == 32 { goto st268 }
	goto st0
st268:
	p++
	if p == pe { goto _test_eof268 }
	fallthrough
case 268:
	switch data[p] {
		case 32: goto st268
		case 50: goto tr340
		case 58: goto tr342
	}
	if data[p] < 51 {
		if 48 <= data[p] && data[p] <= 49 { goto tr339 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto tr343 }
		} else if data[p] >= 65 {
			goto tr343
		}
	} else {
		goto tr341
	}
	goto st0
tr339:
// line 41 "parser.rl"
	{ mark = p }
	goto st269
st269:
	p++
	if p == pe { goto _test_eof269 }
	fallthrough
case 269:
// line 3299 "parser.go"
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st289 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st356 }
	} else {
		goto st356
	}
	goto st0
st270:
	p++
	if p == pe { goto _test_eof270 }
	fallthrough
case 270:
	if data[p] == 50 { goto st287 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st285 }
	} else if data[p] >= 48 {
		goto st271
	}
	goto st0
st271:
	p++
	if p == pe { goto _test_eof271 }
	fallthrough
case 271:
	if data[p] == 46 { goto st272 }
	if 48 <= data[p] && data[p] <= 57 { goto st285 }
	goto st0
st272:
	p++
	if p == pe { goto _test_eof272 }
	fallthrough
case 272:
	if data[p] == 50 { goto st283 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st281 }
	} else if data[p] >= 48 {
		goto st273
	}
	goto st0
st273:
	p++
	if p == pe { goto _test_eof273 }
	fallthrough
case 273:
	if data[p] == 46 { goto st274 }
	if 48 <= data[p] && data[p] <= 57 { goto st281 }
	goto st0
st274:
	p++
	if p == pe { goto _test_eof274 }
	fallthrough
case 274:
	if data[p] == 50 { goto st279 }
	if data[p] > 49 {
		if 51 <= data[p] && data[p] <= 57 { goto st277 }
	} else if data[p] >= 48 {
		goto st275
	}
	goto st0
st275:
	p++
	if p == pe { goto _test_eof275 }
	fallthrough
case 275:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if 48 <= data[p] && data[p] <= 57 { goto st277 }
	goto st0
tr359:
// line 91 "parser.rl"
	{
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	goto st276
tr449:
// line 98 "parser.rl"
	{
		cur.(*cPort).Class = string(data[mark:p])
	}
	goto st276
tr466:
// line 105 "parser.rl"
	{
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	goto st276
tr486:
// line 111 "parser.rl"
	{
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	goto st276
st276:
	p++
	if p == pe { goto _test_eof276 }
	fallthrough
case 276:
// line 3404 "parser.go"
	switch data[p] {
		case 10: goto st258
		case 32: goto st276
	}
	goto st0
st277:
	p++
	if p == pe { goto _test_eof277 }
	fallthrough
case 277:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if 48 <= data[p] && data[p] <= 57 { goto st278 }
	goto st0
st278:
	p++
	if p == pe { goto _test_eof278 }
	fallthrough
case 278:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	goto st0
st279:
	p++
	if p == pe { goto _test_eof279 }
	fallthrough
case 279:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 53: goto st280
	}
	if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st278 }
	} else if data[p] >= 48 {
		goto st277
	}
	goto st0
st280:
	p++
	if p == pe { goto _test_eof280 }
	fallthrough
case 280:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if 48 <= data[p] && data[p] <= 53 { goto st278 }
	goto st0
st281:
	p++
	if p == pe { goto _test_eof281 }
	fallthrough
case 281:
	if data[p] == 46 { goto st274 }
	if 48 <= data[p] && data[p] <= 57 { goto st282 }
	goto st0
st282:
	p++
	if p == pe { goto _test_eof282 }
	fallthrough
case 282:
	if data[p] == 46 { goto st274 }
	goto st0
st283:
	p++
	if p == pe { goto _test_eof283 }
	fallthrough
case 283:
	switch data[p] {
		case 46: goto st274
		case 53: goto st284
	}
	if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st282 }
	} else if data[p] >= 48 {
		goto st281
	}
	goto st0
st284:
	p++
	if p == pe { goto _test_eof284 }
	fallthrough
case 284:
	if data[p] == 46 { goto st274 }
	if 48 <= data[p] && data[p] <= 53 { goto st282 }
	goto st0
st285:
	p++
	if p == pe { goto _test_eof285 }
	fallthrough
case 285:
	if data[p] == 46 { goto st272 }
	if 48 <= data[p] && data[p] <= 57 { goto st286 }
	goto st0
st286:
	p++
	if p == pe { goto _test_eof286 }
	fallthrough
case 286:
	if data[p] == 46 { goto st272 }
	goto st0
st287:
	p++
	if p == pe { goto _test_eof287 }
	fallthrough
case 287:
	switch data[p] {
		case 46: goto st272
		case 53: goto st288
	}
	if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st286 }
	} else if data[p] >= 48 {
		goto st285
	}
	goto st0
st288:
	p++
	if p == pe { goto _test_eof288 }
	fallthrough
case 288:
	if data[p] == 46 { goto st272 }
	if 48 <= data[p] && data[p] <= 53 { goto st286 }
	goto st0
st289:
	p++
	if p == pe { goto _test_eof289 }
	fallthrough
case 289:
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st290 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st355 }
	} else {
		goto st355
	}
	goto st0
st290:
	p++
	if p == pe { goto _test_eof290 }
	fallthrough
case 290:
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st291 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st291 }
	} else {
		goto st291
	}
	goto st0
st291:
	p++
	if p == pe { goto _test_eof291 }
	fallthrough
case 291:
	if data[p] == 58 { goto st292 }
	goto st0
st292:
	p++
	if p == pe { goto _test_eof292 }
	fallthrough
case 292:
	if data[p] == 58 { goto st349 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st293 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st293 }
	} else {
		goto st293
	}
	goto st0
st293:
	p++
	if p == pe { goto _test_eof293 }
	fallthrough
case 293:
	if data[p] == 58 { goto st297 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st294 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st294 }
	} else {
		goto st294
	}
	goto st0
st294:
	p++
	if p == pe { goto _test_eof294 }
	fallthrough
case 294:
	if data[p] == 58 { goto st297 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st295 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st295 }
	} else {
		goto st295
	}
	goto st0
st295:
	p++
	if p == pe { goto _test_eof295 }
	fallthrough
case 295:
	if data[p] == 58 { goto st297 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st296 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st296 }
	} else {
		goto st296
	}
	goto st0
st296:
	p++
	if p == pe { goto _test_eof296 }
	fallthrough
case 296:
	if data[p] == 58 { goto st297 }
	goto st0
st297:
	p++
	if p == pe { goto _test_eof297 }
	fallthrough
case 297:
	if data[p] == 58 { goto st332 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st298 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st298 }
	} else {
		goto st298
	}
	goto st0
st298:
	p++
	if p == pe { goto _test_eof298 }
	fallthrough
case 298:
	if data[p] == 58 { goto st302 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st299 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st299 }
	} else {
		goto st299
	}
	goto st0
st299:
	p++
	if p == pe { goto _test_eof299 }
	fallthrough
case 299:
	if data[p] == 58 { goto st302 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st300 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st300 }
	} else {
		goto st300
	}
	goto st0
st300:
	p++
	if p == pe { goto _test_eof300 }
	fallthrough
case 300:
	if data[p] == 58 { goto st302 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st301 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st301 }
	} else {
		goto st301
	}
	goto st0
st301:
	p++
	if p == pe { goto _test_eof301 }
	fallthrough
case 301:
	if data[p] == 58 { goto st302 }
	goto st0
st302:
	p++
	if p == pe { goto _test_eof302 }
	fallthrough
case 302:
	if data[p] == 58 { goto st348 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st303 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st303 }
	} else {
		goto st303
	}
	goto st0
st303:
	p++
	if p == pe { goto _test_eof303 }
	fallthrough
case 303:
	if data[p] == 58 { goto st307 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st304 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st304 }
	} else {
		goto st304
	}
	goto st0
st304:
	p++
	if p == pe { goto _test_eof304 }
	fallthrough
case 304:
	if data[p] == 58 { goto st307 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st305 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st305 }
	} else {
		goto st305
	}
	goto st0
st305:
	p++
	if p == pe { goto _test_eof305 }
	fallthrough
case 305:
	if data[p] == 58 { goto st307 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st306 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st306 }
	} else {
		goto st306
	}
	goto st0
st306:
	p++
	if p == pe { goto _test_eof306 }
	fallthrough
case 306:
	if data[p] == 58 { goto st307 }
	goto st0
st307:
	p++
	if p == pe { goto _test_eof307 }
	fallthrough
case 307:
	if data[p] == 58 { goto st332 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st308 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st308 }
	} else {
		goto st308
	}
	goto st0
st308:
	p++
	if p == pe { goto _test_eof308 }
	fallthrough
case 308:
	if data[p] == 58 { goto st312 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st309 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st309 }
	} else {
		goto st309
	}
	goto st0
st309:
	p++
	if p == pe { goto _test_eof309 }
	fallthrough
case 309:
	if data[p] == 58 { goto st312 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st310 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st310 }
	} else {
		goto st310
	}
	goto st0
st310:
	p++
	if p == pe { goto _test_eof310 }
	fallthrough
case 310:
	if data[p] == 58 { goto st312 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st311 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st311 }
	} else {
		goto st311
	}
	goto st0
st311:
	p++
	if p == pe { goto _test_eof311 }
	fallthrough
case 311:
	if data[p] == 58 { goto st312 }
	goto st0
st312:
	p++
	if p == pe { goto _test_eof312 }
	fallthrough
case 312:
	if data[p] == 58 { goto st327 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st313 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st313 }
	} else {
		goto st313
	}
	goto st0
st313:
	p++
	if p == pe { goto _test_eof313 }
	fallthrough
case 313:
	if data[p] == 58 { goto st317 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st314 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st314 }
	} else {
		goto st314
	}
	goto st0
st314:
	p++
	if p == pe { goto _test_eof314 }
	fallthrough
case 314:
	if data[p] == 58 { goto st317 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st315 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st315 }
	} else {
		goto st315
	}
	goto st0
st315:
	p++
	if p == pe { goto _test_eof315 }
	fallthrough
case 315:
	if data[p] == 58 { goto st317 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st316 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st316 }
	} else {
		goto st316
	}
	goto st0
st316:
	p++
	if p == pe { goto _test_eof316 }
	fallthrough
case 316:
	if data[p] == 58 { goto st317 }
	goto st0
st317:
	p++
	if p == pe { goto _test_eof317 }
	fallthrough
case 317:
	if data[p] == 58 { goto st326 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st318 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st318 }
	} else {
		goto st318
	}
	goto st0
st318:
	p++
	if p == pe { goto _test_eof318 }
	fallthrough
case 318:
	if data[p] == 58 { goto st322 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st319 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st319 }
	} else {
		goto st319
	}
	goto st0
st319:
	p++
	if p == pe { goto _test_eof319 }
	fallthrough
case 319:
	if data[p] == 58 { goto st322 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st320 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st320 }
	} else {
		goto st320
	}
	goto st0
st320:
	p++
	if p == pe { goto _test_eof320 }
	fallthrough
case 320:
	if data[p] == 58 { goto st322 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st321 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st321 }
	} else {
		goto st321
	}
	goto st0
st321:
	p++
	if p == pe { goto _test_eof321 }
	fallthrough
case 321:
	if data[p] == 58 { goto st322 }
	goto st0
st322:
	p++
	if p == pe { goto _test_eof322 }
	fallthrough
case 322:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st323 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st323 }
	} else {
		goto st323
	}
	goto st0
st323:
	p++
	if p == pe { goto _test_eof323 }
	fallthrough
case 323:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st324 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st324 }
	} else {
		goto st324
	}
	goto st0
st324:
	p++
	if p == pe { goto _test_eof324 }
	fallthrough
case 324:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st325 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st325 }
	} else {
		goto st325
	}
	goto st0
st325:
	p++
	if p == pe { goto _test_eof325 }
	fallthrough
case 325:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st278 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st278 }
	} else {
		goto st278
	}
	goto st0
st326:
	p++
	if p == pe { goto _test_eof326 }
	fallthrough
case 326:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st323 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st323 }
	} else {
		goto st323
	}
	goto st0
st327:
	p++
	if p == pe { goto _test_eof327 }
	fallthrough
case 327:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st328 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st328 }
	} else {
		goto st328
	}
	goto st0
st328:
	p++
	if p == pe { goto _test_eof328 }
	fallthrough
case 328:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st322
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st329 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st329 }
	} else {
		goto st329
	}
	goto st0
st329:
	p++
	if p == pe { goto _test_eof329 }
	fallthrough
case 329:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st322
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st330 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st330 }
	} else {
		goto st330
	}
	goto st0
st330:
	p++
	if p == pe { goto _test_eof330 }
	fallthrough
case 330:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st322
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st331 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st331 }
	} else {
		goto st331
	}
	goto st0
st331:
	p++
	if p == pe { goto _test_eof331 }
	fallthrough
case 331:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st322
	}
	goto st0
st332:
	p++
	if p == pe { goto _test_eof332 }
	fallthrough
case 332:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st333 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st333 }
	} else {
		goto st333
	}
	goto st0
st333:
	p++
	if p == pe { goto _test_eof333 }
	fallthrough
case 333:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st337
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st334 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st334 }
	} else {
		goto st334
	}
	goto st0
st334:
	p++
	if p == pe { goto _test_eof334 }
	fallthrough
case 334:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st337
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st335 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st335 }
	} else {
		goto st335
	}
	goto st0
st335:
	p++
	if p == pe { goto _test_eof335 }
	fallthrough
case 335:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st337
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st336 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st336 }
	} else {
		goto st336
	}
	goto st0
st336:
	p++
	if p == pe { goto _test_eof336 }
	fallthrough
case 336:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st337
	}
	goto st0
st337:
	p++
	if p == pe { goto _test_eof337 }
	fallthrough
case 337:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st338 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st338 }
	} else {
		goto st338
	}
	goto st0
st338:
	p++
	if p == pe { goto _test_eof338 }
	fallthrough
case 338:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st342
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st339 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st339 }
	} else {
		goto st339
	}
	goto st0
st339:
	p++
	if p == pe { goto _test_eof339 }
	fallthrough
case 339:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st342
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st340 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st340 }
	} else {
		goto st340
	}
	goto st0
st340:
	p++
	if p == pe { goto _test_eof340 }
	fallthrough
case 340:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st342
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st341 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st341 }
	} else {
		goto st341
	}
	goto st0
st341:
	p++
	if p == pe { goto _test_eof341 }
	fallthrough
case 341:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st342
	}
	goto st0
st342:
	p++
	if p == pe { goto _test_eof342 }
	fallthrough
case 342:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st343 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st343 }
	} else {
		goto st343
	}
	goto st0
st343:
	p++
	if p == pe { goto _test_eof343 }
	fallthrough
case 343:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st347
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st344 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st344 }
	} else {
		goto st344
	}
	goto st0
st344:
	p++
	if p == pe { goto _test_eof344 }
	fallthrough
case 344:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st347
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st345 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st345 }
	} else {
		goto st345
	}
	goto st0
st345:
	p++
	if p == pe { goto _test_eof345 }
	fallthrough
case 345:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st347
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st346 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st346 }
	} else {
		goto st346
	}
	goto st0
st346:
	p++
	if p == pe { goto _test_eof346 }
	fallthrough
case 346:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st347
	}
	goto st0
st347:
	p++
	if p == pe { goto _test_eof347 }
	fallthrough
case 347:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st328 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st328 }
	} else {
		goto st328
	}
	goto st0
st348:
	p++
	if p == pe { goto _test_eof348 }
	fallthrough
case 348:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st338 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st338 }
	} else {
		goto st338
	}
	goto st0
st349:
	p++
	if p == pe { goto _test_eof349 }
	fallthrough
case 349:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st350 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st350 }
	} else {
		goto st350
	}
	goto st0
st350:
	p++
	if p == pe { goto _test_eof350 }
	fallthrough
case 350:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st354
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st351 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st351 }
	} else {
		goto st351
	}
	goto st0
st351:
	p++
	if p == pe { goto _test_eof351 }
	fallthrough
case 351:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st354
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st352 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st352 }
	} else {
		goto st352
	}
	goto st0
st352:
	p++
	if p == pe { goto _test_eof352 }
	fallthrough
case 352:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st354
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st353 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st353 }
	} else {
		goto st353
	}
	goto st0
st353:
	p++
	if p == pe { goto _test_eof353 }
	fallthrough
case 353:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr359
		case 58: goto st354
	}
	goto st0
st354:
	p++
	if p == pe { goto _test_eof354 }
	fallthrough
case 354:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st333 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st333 }
	} else {
		goto st333
	}
	goto st0
st355:
	p++
	if p == pe { goto _test_eof355 }
	fallthrough
case 355:
	if data[p] == 58 { goto st292 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st291 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st291 }
	} else {
		goto st291
	}
	goto st0
st356:
	p++
	if p == pe { goto _test_eof356 }
	fallthrough
case 356:
	if data[p] == 58 { goto st292 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st355 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st355 }
	} else {
		goto st355
	}
	goto st0
tr340:
// line 41 "parser.rl"
	{ mark = p }
	goto st357
st357:
	p++
	if p == pe { goto _test_eof357 }
	fallthrough
case 357:
// line 4516 "parser.go"
	switch data[p] {
		case 46: goto st270
		case 53: goto st358
		case 58: goto st292
	}
	if data[p] < 54 {
		if 48 <= data[p] && data[p] <= 52 { goto st289 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st356 }
		} else if data[p] >= 65 {
			goto st356
		}
	} else {
		goto st359
	}
	goto st0
st358:
	p++
	if p == pe { goto _test_eof358 }
	fallthrough
case 358:
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 54 {
		if 48 <= data[p] && data[p] <= 53 { goto st290 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st355 }
		} else if data[p] >= 65 {
			goto st355
		}
	} else {
		goto st355
	}
	goto st0
st359:
	p++
	if p == pe { goto _test_eof359 }
	fallthrough
case 359:
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st355 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st355 }
	} else {
		goto st355
	}
	goto st0
tr341:
// line 41 "parser.rl"
	{ mark = p }
	goto st360
st360:
	p++
	if p == pe { goto _test_eof360 }
	fallthrough
case 360:
// line 4581 "parser.go"
	switch data[p] {
		case 46: goto st270
		case 58: goto st292
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st359 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st356 }
	} else {
		goto st356
	}
	goto st0
tr342:
// line 41 "parser.rl"
	{ mark = p }
	goto st361
st361:
	p++
	if p == pe { goto _test_eof361 }
	fallthrough
case 361:
// line 4603 "parser.go"
	if data[p] == 58 { goto st349 }
	goto st0
tr343:
// line 41 "parser.rl"
	{ mark = p }
	goto st362
st362:
	p++
	if p == pe { goto _test_eof362 }
	fallthrough
case 362:
// line 4615 "parser.go"
	if data[p] == 58 { goto st292 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st356 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st356 }
	} else {
		goto st356
	}
	goto st0
st363:
	p++
	if p == pe { goto _test_eof363 }
	fallthrough
case 363:
	if data[p] == 108 { goto st364 }
	goto st0
st364:
	p++
	if p == pe { goto _test_eof364 }
	fallthrough
case 364:
	if data[p] == 97 { goto st365 }
	goto st0
st365:
	p++
	if p == pe { goto _test_eof365 }
	fallthrough
case 365:
	if data[p] == 115 { goto st366 }
	goto st0
st366:
	p++
	if p == pe { goto _test_eof366 }
	fallthrough
case 366:
	if data[p] == 115 { goto st367 }
	goto st0
st367:
	p++
	if p == pe { goto _test_eof367 }
	fallthrough
case 367:
	if data[p] == 32 { goto st368 }
	goto st0
st368:
	p++
	if p == pe { goto _test_eof368 }
	fallthrough
case 368:
	switch data[p] {
		case 32: goto st368
		case 61: goto st369
	}
	goto st0
st369:
	p++
	if p == pe { goto _test_eof369 }
	fallthrough
case 369:
	if data[p] == 32 { goto st370 }
	goto st0
st370:
	p++
	if p == pe { goto _test_eof370 }
	fallthrough
case 370:
	switch data[p] {
		case 32: goto st370
		case 99: goto tr441
		case 115: goto tr442
	}
	goto st0
tr441:
// line 41 "parser.rl"
	{ mark = p }
	goto st371
st371:
	p++
	if p == pe { goto _test_eof371 }
	fallthrough
case 371:
// line 4697 "parser.go"
	if data[p] == 108 { goto st372 }
	goto st0
st372:
	p++
	if p == pe { goto _test_eof372 }
	fallthrough
case 372:
	if data[p] == 105 { goto st373 }
	goto st0
st373:
	p++
	if p == pe { goto _test_eof373 }
	fallthrough
case 373:
	if data[p] == 101 { goto st374 }
	goto st0
st374:
	p++
	if p == pe { goto _test_eof374 }
	fallthrough
case 374:
	if data[p] == 110 { goto st375 }
	goto st0
st375:
	p++
	if p == pe { goto _test_eof375 }
	fallthrough
case 375:
	if data[p] == 116 { goto st376 }
	goto st0
st376:
	p++
	if p == pe { goto _test_eof376 }
	fallthrough
case 376:
	switch data[p] {
		case 10: goto tr448
		case 32: goto tr449
	}
	goto st0
tr442:
// line 41 "parser.rl"
	{ mark = p }
	goto st377
st377:
	p++
	if p == pe { goto _test_eof377 }
	fallthrough
case 377:
// line 4747 "parser.go"
	if data[p] == 101 { goto st378 }
	goto st0
st378:
	p++
	if p == pe { goto _test_eof378 }
	fallthrough
case 378:
	if data[p] == 114 { goto st379 }
	goto st0
st379:
	p++
	if p == pe { goto _test_eof379 }
	fallthrough
case 379:
	if data[p] == 118 { goto st380 }
	goto st0
st380:
	p++
	if p == pe { goto _test_eof380 }
	fallthrough
case 380:
	if data[p] == 101 { goto st381 }
	goto st0
st381:
	p++
	if p == pe { goto _test_eof381 }
	fallthrough
case 381:
	if data[p] == 114 { goto st376 }
	goto st0
st382:
	p++
	if p == pe { goto _test_eof382 }
	fallthrough
case 382:
	if data[p] == 115 { goto st383 }
	goto st0
st383:
	p++
	if p == pe { goto _test_eof383 }
	fallthrough
case 383:
	if data[p] == 108 { goto st384 }
	goto st0
st384:
	p++
	if p == pe { goto _test_eof384 }
	fallthrough
case 384:
	if data[p] == 32 { goto st385 }
	goto st0
st385:
	p++
	if p == pe { goto _test_eof385 }
	fallthrough
case 385:
	switch data[p] {
		case 32: goto st385
		case 61: goto st386
	}
	goto st0
st386:
	p++
	if p == pe { goto _test_eof386 }
	fallthrough
case 386:
	if data[p] == 32 { goto st387 }
	goto st0
st387:
	p++
	if p == pe { goto _test_eof387 }
	fallthrough
case 387:
	switch data[p] {
		case 32: goto st387
		case 102: goto tr460
		case 110: goto tr461
		case 111: goto tr462
		case 116: goto tr463
		case 121: goto tr464
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr459 }
	goto st0
tr459:
// line 41 "parser.rl"
	{ mark = p }
	goto st388
st388:
	p++
	if p == pe { goto _test_eof388 }
	fallthrough
case 388:
// line 4840 "parser.go"
	switch data[p] {
		case 10: goto tr465
		case 32: goto tr466
	}
	goto st0
tr460:
// line 41 "parser.rl"
	{ mark = p }
	goto st389
st389:
	p++
	if p == pe { goto _test_eof389 }
	fallthrough
case 389:
// line 4855 "parser.go"
	if data[p] == 97 { goto st390 }
	goto st0
st390:
	p++
	if p == pe { goto _test_eof390 }
	fallthrough
case 390:
	if data[p] == 108 { goto st391 }
	goto st0
st391:
	p++
	if p == pe { goto _test_eof391 }
	fallthrough
case 391:
	if data[p] == 115 { goto st392 }
	goto st0
st392:
	p++
	if p == pe { goto _test_eof392 }
	fallthrough
case 392:
	if data[p] == 101 { goto st388 }
	goto st0
tr461:
// line 41 "parser.rl"
	{ mark = p }
	goto st393
st393:
	p++
	if p == pe { goto _test_eof393 }
	fallthrough
case 393:
// line 4888 "parser.go"
	if data[p] == 111 { goto st388 }
	goto st0
tr462:
// line 41 "parser.rl"
	{ mark = p }
	goto st394
st394:
	p++
	if p == pe { goto _test_eof394 }
	fallthrough
case 394:
// line 4900 "parser.go"
	switch data[p] {
		case 102: goto st395
		case 110: goto st388
	}
	goto st0
st395:
	p++
	if p == pe { goto _test_eof395 }
	fallthrough
case 395:
	if data[p] == 102 { goto st388 }
	goto st0
tr463:
// line 41 "parser.rl"
	{ mark = p }
	goto st396
st396:
	p++
	if p == pe { goto _test_eof396 }
	fallthrough
case 396:
// line 4922 "parser.go"
	if data[p] == 114 { goto st397 }
	goto st0
st397:
	p++
	if p == pe { goto _test_eof397 }
	fallthrough
case 397:
	if data[p] == 117 { goto st392 }
	goto st0
tr464:
// line 41 "parser.rl"
	{ mark = p }
	goto st398
st398:
	p++
	if p == pe { goto _test_eof398 }
	fallthrough
case 398:
// line 4941 "parser.go"
	if data[p] == 101 { goto st399 }
	goto st0
st399:
	p++
	if p == pe { goto _test_eof399 }
	fallthrough
case 399:
	if data[p] == 115 { goto st388 }
	goto st0
st400:
	p++
	if p == pe { goto _test_eof400 }
	fallthrough
case 400:
	if data[p] == 105 { goto st401 }
	goto st0
st401:
	p++
	if p == pe { goto _test_eof401 }
	fallthrough
case 401:
	if data[p] == 112 { goto st402 }
	goto st0
st402:
	p++
	if p == pe { goto _test_eof402 }
	fallthrough
case 402:
	if data[p] == 32 { goto st403 }
	goto st0
st403:
	p++
	if p == pe { goto _test_eof403 }
	fallthrough
case 403:
	switch data[p] {
		case 32: goto st403
		case 61: goto st404
	}
	goto st0
st404:
	p++
	if p == pe { goto _test_eof404 }
	fallthrough
case 404:
	if data[p] == 32 { goto st405 }
	goto st0
st405:
	p++
	if p == pe { goto _test_eof405 }
	fallthrough
case 405:
	switch data[p] {
		case 32: goto st405
		case 102: goto tr480
		case 110: goto tr481
		case 111: goto tr482
		case 116: goto tr483
		case 121: goto tr484
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr479 }
	goto st0
tr479:
// line 41 "parser.rl"
	{ mark = p }
	goto st406
st406:
	p++
	if p == pe { goto _test_eof406 }
	fallthrough
case 406:
// line 5013 "parser.go"
	switch data[p] {
		case 10: goto tr485
		case 32: goto tr486
	}
	goto st0
tr480:
// line 41 "parser.rl"
	{ mark = p }
	goto st407
st407:
	p++
	if p == pe { goto _test_eof407 }
	fallthrough
case 407:
// line 5028 "parser.go"
	if data[p] == 97 { goto st408 }
	goto st0
st408:
	p++
	if p == pe { goto _test_eof408 }
	fallthrough
case 408:
	if data[p] == 108 { goto st409 }
	goto st0
st409:
	p++
	if p == pe { goto _test_eof409 }
	fallthrough
case 409:
	if data[p] == 115 { goto st410 }
	goto st0
st410:
	p++
	if p == pe { goto _test_eof410 }
	fallthrough
case 410:
	if data[p] == 101 { goto st406 }
	goto st0
tr481:
// line 41 "parser.rl"
	{ mark = p }
	goto st411
st411:
	p++
	if p == pe { goto _test_eof411 }
	fallthrough
case 411:
// line 5061 "parser.go"
	if data[p] == 111 { goto st406 }
	goto st0
tr482:
// line 41 "parser.rl"
	{ mark = p }
	goto st412
st412:
	p++
	if p == pe { goto _test_eof412 }
	fallthrough
case 412:
// line 5073 "parser.go"
	switch data[p] {
		case 102: goto st413
		case 110: goto st406
	}
	goto st0
st413:
	p++
	if p == pe { goto _test_eof413 }
	fallthrough
case 413:
	if data[p] == 102 { goto st406 }
	goto st0
tr483:
// line 41 "parser.rl"
	{ mark = p }
	goto st414
st414:
	p++
	if p == pe { goto _test_eof414 }
	fallthrough
case 414:
// line 5095 "parser.go"
	if data[p] == 114 { goto st415 }
	goto st0
st415:
	p++
	if p == pe { goto _test_eof415 }
	fallthrough
case 415:
	if data[p] == 117 { goto st410 }
	goto st0
tr484:
// line 41 "parser.rl"
	{ mark = p }
	goto st416
st416:
	p++
	if p == pe { goto _test_eof416 }
	fallthrough
case 416:
// line 5114 "parser.go"
	if data[p] == 101 { goto st417 }
	goto st0
st417:
	p++
	if p == pe { goto _test_eof417 }
	fallthrough
case 417:
	if data[p] == 115 { goto st406 }
	goto st0
tr322:
// line 91 "parser.rl"
	{
		cur.(*cPort).BindIP = net.ParseIP(string(data[mark:p]))
	}
	goto st418
tr583:
// line 98 "parser.rl"
	{
		cur.(*cPort).Class = string(data[mark:p])
	}
	goto st418
tr601:
// line 105 "parser.rl"
	{
		cur.(*cPort).SSL = getbool(data[mark:p])
	}
	goto st418
tr622:
// line 111 "parser.rl"
	{
		cur.(*cPort).Zip = getbool(data[mark:p])
	}
	goto st418
st418:
	p++
	if p == pe { goto _test_eof418 }
	fallthrough
case 418:
// line 5153 "parser.go"
	switch data[p] {
		case 10: goto st258
		case 32: goto st418
		case 125: goto st560
	}
	if 9 <= data[p] && data[p] <= 13 { goto st257 }
	goto st0
st419:
	p++
	if p == pe { goto _test_eof419 }
	fallthrough
case 419:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] > 13 {
		if 48 <= data[p] && data[p] <= 57 { goto st420 }
	} else if data[p] >= 9 {
		goto tr320
	}
	goto st0
st420:
	p++
	if p == pe { goto _test_eof420 }
	fallthrough
case 420:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st421:
	p++
	if p == pe { goto _test_eof421 }
	fallthrough
case 421:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 53: goto st422
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st420 }
	} else {
		goto st419
	}
	goto st0
st422:
	p++
	if p == pe { goto _test_eof422 }
	fallthrough
case 422:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] > 13 {
		if 48 <= data[p] && data[p] <= 53 { goto st420 }
	} else if data[p] >= 9 {
		goto tr320
	}
	goto st0
st423:
	p++
	if p == pe { goto _test_eof423 }
	fallthrough
case 423:
	if data[p] == 46 { goto st255 }
	if 48 <= data[p] && data[p] <= 57 { goto st424 }
	goto st0
st424:
	p++
	if p == pe { goto _test_eof424 }
	fallthrough
case 424:
	if data[p] == 46 { goto st255 }
	goto st0
st425:
	p++
	if p == pe { goto _test_eof425 }
	fallthrough
case 425:
	switch data[p] {
		case 46: goto st255
		case 53: goto st426
	}
	if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st424 }
	} else if data[p] >= 48 {
		goto st423
	}
	goto st0
st426:
	p++
	if p == pe { goto _test_eof426 }
	fallthrough
case 426:
	if data[p] == 46 { goto st255 }
	if 48 <= data[p] && data[p] <= 53 { goto st424 }
	goto st0
st427:
	p++
	if p == pe { goto _test_eof427 }
	fallthrough
case 427:
	if data[p] == 46 { goto st253 }
	if 48 <= data[p] && data[p] <= 57 { goto st428 }
	goto st0
st428:
	p++
	if p == pe { goto _test_eof428 }
	fallthrough
case 428:
	if data[p] == 46 { goto st253 }
	goto st0
st429:
	p++
	if p == pe { goto _test_eof429 }
	fallthrough
case 429:
	switch data[p] {
		case 46: goto st253
		case 53: goto st430
	}
	if data[p] > 52 {
		if 54 <= data[p] && data[p] <= 57 { goto st428 }
	} else if data[p] >= 48 {
		goto st427
	}
	goto st0
st430:
	p++
	if p == pe { goto _test_eof430 }
	fallthrough
case 430:
	if data[p] == 46 { goto st253 }
	if 48 <= data[p] && data[p] <= 53 { goto st428 }
	goto st0
st431:
	p++
	if p == pe { goto _test_eof431 }
	fallthrough
case 431:
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st432 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st497 }
	} else {
		goto st497
	}
	goto st0
st432:
	p++
	if p == pe { goto _test_eof432 }
	fallthrough
case 432:
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st433 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st433 }
	} else {
		goto st433
	}
	goto st0
st433:
	p++
	if p == pe { goto _test_eof433 }
	fallthrough
case 433:
	if data[p] == 58 { goto st434 }
	goto st0
st434:
	p++
	if p == pe { goto _test_eof434 }
	fallthrough
case 434:
	if data[p] == 58 { goto st491 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st435 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st435 }
	} else {
		goto st435
	}
	goto st0
st435:
	p++
	if p == pe { goto _test_eof435 }
	fallthrough
case 435:
	if data[p] == 58 { goto st439 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st436 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st436 }
	} else {
		goto st436
	}
	goto st0
st436:
	p++
	if p == pe { goto _test_eof436 }
	fallthrough
case 436:
	if data[p] == 58 { goto st439 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st437 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st437 }
	} else {
		goto st437
	}
	goto st0
st437:
	p++
	if p == pe { goto _test_eof437 }
	fallthrough
case 437:
	if data[p] == 58 { goto st439 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st438 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st438 }
	} else {
		goto st438
	}
	goto st0
st438:
	p++
	if p == pe { goto _test_eof438 }
	fallthrough
case 438:
	if data[p] == 58 { goto st439 }
	goto st0
st439:
	p++
	if p == pe { goto _test_eof439 }
	fallthrough
case 439:
	if data[p] == 58 { goto st474 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st440 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st440 }
	} else {
		goto st440
	}
	goto st0
st440:
	p++
	if p == pe { goto _test_eof440 }
	fallthrough
case 440:
	if data[p] == 58 { goto st444 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st441 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st441 }
	} else {
		goto st441
	}
	goto st0
st441:
	p++
	if p == pe { goto _test_eof441 }
	fallthrough
case 441:
	if data[p] == 58 { goto st444 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st442 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st442 }
	} else {
		goto st442
	}
	goto st0
st442:
	p++
	if p == pe { goto _test_eof442 }
	fallthrough
case 442:
	if data[p] == 58 { goto st444 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st443 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st443 }
	} else {
		goto st443
	}
	goto st0
st443:
	p++
	if p == pe { goto _test_eof443 }
	fallthrough
case 443:
	if data[p] == 58 { goto st444 }
	goto st0
st444:
	p++
	if p == pe { goto _test_eof444 }
	fallthrough
case 444:
	if data[p] == 58 { goto st490 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st445 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st445 }
	} else {
		goto st445
	}
	goto st0
st445:
	p++
	if p == pe { goto _test_eof445 }
	fallthrough
case 445:
	if data[p] == 58 { goto st449 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st446 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st446 }
	} else {
		goto st446
	}
	goto st0
st446:
	p++
	if p == pe { goto _test_eof446 }
	fallthrough
case 446:
	if data[p] == 58 { goto st449 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st447 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st447 }
	} else {
		goto st447
	}
	goto st0
st447:
	p++
	if p == pe { goto _test_eof447 }
	fallthrough
case 447:
	if data[p] == 58 { goto st449 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st448 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st448 }
	} else {
		goto st448
	}
	goto st0
st448:
	p++
	if p == pe { goto _test_eof448 }
	fallthrough
case 448:
	if data[p] == 58 { goto st449 }
	goto st0
st449:
	p++
	if p == pe { goto _test_eof449 }
	fallthrough
case 449:
	if data[p] == 58 { goto st474 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st450 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st450 }
	} else {
		goto st450
	}
	goto st0
st450:
	p++
	if p == pe { goto _test_eof450 }
	fallthrough
case 450:
	if data[p] == 58 { goto st454 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st451 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st451 }
	} else {
		goto st451
	}
	goto st0
st451:
	p++
	if p == pe { goto _test_eof451 }
	fallthrough
case 451:
	if data[p] == 58 { goto st454 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st452 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st452 }
	} else {
		goto st452
	}
	goto st0
st452:
	p++
	if p == pe { goto _test_eof452 }
	fallthrough
case 452:
	if data[p] == 58 { goto st454 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st453 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st453 }
	} else {
		goto st453
	}
	goto st0
st453:
	p++
	if p == pe { goto _test_eof453 }
	fallthrough
case 453:
	if data[p] == 58 { goto st454 }
	goto st0
st454:
	p++
	if p == pe { goto _test_eof454 }
	fallthrough
case 454:
	if data[p] == 58 { goto st469 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st455 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st455 }
	} else {
		goto st455
	}
	goto st0
st455:
	p++
	if p == pe { goto _test_eof455 }
	fallthrough
case 455:
	if data[p] == 58 { goto st459 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st456 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st456 }
	} else {
		goto st456
	}
	goto st0
st456:
	p++
	if p == pe { goto _test_eof456 }
	fallthrough
case 456:
	if data[p] == 58 { goto st459 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st457 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st457 }
	} else {
		goto st457
	}
	goto st0
st457:
	p++
	if p == pe { goto _test_eof457 }
	fallthrough
case 457:
	if data[p] == 58 { goto st459 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st458 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st458 }
	} else {
		goto st458
	}
	goto st0
st458:
	p++
	if p == pe { goto _test_eof458 }
	fallthrough
case 458:
	if data[p] == 58 { goto st459 }
	goto st0
st459:
	p++
	if p == pe { goto _test_eof459 }
	fallthrough
case 459:
	if data[p] == 58 { goto st468 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st460 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st460 }
	} else {
		goto st460
	}
	goto st0
st460:
	p++
	if p == pe { goto _test_eof460 }
	fallthrough
case 460:
	if data[p] == 58 { goto st464 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st461 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st461 }
	} else {
		goto st461
	}
	goto st0
st461:
	p++
	if p == pe { goto _test_eof461 }
	fallthrough
case 461:
	if data[p] == 58 { goto st464 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st462 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st462 }
	} else {
		goto st462
	}
	goto st0
st462:
	p++
	if p == pe { goto _test_eof462 }
	fallthrough
case 462:
	if data[p] == 58 { goto st464 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st463 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st463 }
	} else {
		goto st463
	}
	goto st0
st463:
	p++
	if p == pe { goto _test_eof463 }
	fallthrough
case 463:
	if data[p] == 58 { goto st464 }
	goto st0
st464:
	p++
	if p == pe { goto _test_eof464 }
	fallthrough
case 464:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st465 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st465 }
	} else {
		goto st465
	}
	goto st0
st465:
	p++
	if p == pe { goto _test_eof465 }
	fallthrough
case 465:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st466 }
		} else if data[p] >= 65 {
			goto st466
		}
	} else {
		goto st466
	}
	goto st0
st466:
	p++
	if p == pe { goto _test_eof466 }
	fallthrough
case 466:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st467 }
		} else if data[p] >= 65 {
			goto st467
		}
	} else {
		goto st467
	}
	goto st0
st467:
	p++
	if p == pe { goto _test_eof467 }
	fallthrough
case 467:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st420 }
		} else if data[p] >= 65 {
			goto st420
		}
	} else {
		goto st420
	}
	goto st0
st468:
	p++
	if p == pe { goto _test_eof468 }
	fallthrough
case 468:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st465 }
		} else if data[p] >= 65 {
			goto st465
		}
	} else {
		goto st465
	}
	goto st0
st469:
	p++
	if p == pe { goto _test_eof469 }
	fallthrough
case 469:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st470 }
		} else if data[p] >= 65 {
			goto st470
		}
	} else {
		goto st470
	}
	goto st0
st470:
	p++
	if p == pe { goto _test_eof470 }
	fallthrough
case 470:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st464
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st471 }
		} else if data[p] >= 65 {
			goto st471
		}
	} else {
		goto st471
	}
	goto st0
st471:
	p++
	if p == pe { goto _test_eof471 }
	fallthrough
case 471:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st464
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st472 }
		} else if data[p] >= 65 {
			goto st472
		}
	} else {
		goto st472
	}
	goto st0
st472:
	p++
	if p == pe { goto _test_eof472 }
	fallthrough
case 472:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st464
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st473 }
		} else if data[p] >= 65 {
			goto st473
		}
	} else {
		goto st473
	}
	goto st0
st473:
	p++
	if p == pe { goto _test_eof473 }
	fallthrough
case 473:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st464
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st474:
	p++
	if p == pe { goto _test_eof474 }
	fallthrough
case 474:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st475 }
		} else if data[p] >= 65 {
			goto st475
		}
	} else {
		goto st475
	}
	goto st0
st475:
	p++
	if p == pe { goto _test_eof475 }
	fallthrough
case 475:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st479
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st476 }
		} else if data[p] >= 65 {
			goto st476
		}
	} else {
		goto st476
	}
	goto st0
st476:
	p++
	if p == pe { goto _test_eof476 }
	fallthrough
case 476:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st479
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st477 }
		} else if data[p] >= 65 {
			goto st477
		}
	} else {
		goto st477
	}
	goto st0
st477:
	p++
	if p == pe { goto _test_eof477 }
	fallthrough
case 477:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st479
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st478 }
		} else if data[p] >= 65 {
			goto st478
		}
	} else {
		goto st478
	}
	goto st0
st478:
	p++
	if p == pe { goto _test_eof478 }
	fallthrough
case 478:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st479
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st479:
	p++
	if p == pe { goto _test_eof479 }
	fallthrough
case 479:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st480 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st480 }
	} else {
		goto st480
	}
	goto st0
st480:
	p++
	if p == pe { goto _test_eof480 }
	fallthrough
case 480:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st484
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st481 }
		} else if data[p] >= 65 {
			goto st481
		}
	} else {
		goto st481
	}
	goto st0
st481:
	p++
	if p == pe { goto _test_eof481 }
	fallthrough
case 481:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st484
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st482 }
		} else if data[p] >= 65 {
			goto st482
		}
	} else {
		goto st482
	}
	goto st0
st482:
	p++
	if p == pe { goto _test_eof482 }
	fallthrough
case 482:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st484
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st483 }
		} else if data[p] >= 65 {
			goto st483
		}
	} else {
		goto st483
	}
	goto st0
st483:
	p++
	if p == pe { goto _test_eof483 }
	fallthrough
case 483:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st484
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st484:
	p++
	if p == pe { goto _test_eof484 }
	fallthrough
case 484:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st485 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st485 }
	} else {
		goto st485
	}
	goto st0
st485:
	p++
	if p == pe { goto _test_eof485 }
	fallthrough
case 485:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st489
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st486 }
		} else if data[p] >= 65 {
			goto st486
		}
	} else {
		goto st486
	}
	goto st0
st486:
	p++
	if p == pe { goto _test_eof486 }
	fallthrough
case 486:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st489
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st487 }
		} else if data[p] >= 65 {
			goto st487
		}
	} else {
		goto st487
	}
	goto st0
st487:
	p++
	if p == pe { goto _test_eof487 }
	fallthrough
case 487:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st489
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st488 }
		} else if data[p] >= 65 {
			goto st488
		}
	} else {
		goto st488
	}
	goto st0
st488:
	p++
	if p == pe { goto _test_eof488 }
	fallthrough
case 488:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st489
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st489:
	p++
	if p == pe { goto _test_eof489 }
	fallthrough
case 489:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st470 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st470 }
	} else {
		goto st470
	}
	goto st0
st490:
	p++
	if p == pe { goto _test_eof490 }
	fallthrough
case 490:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st480 }
		} else if data[p] >= 65 {
			goto st480
		}
	} else {
		goto st480
	}
	goto st0
st491:
	p++
	if p == pe { goto _test_eof491 }
	fallthrough
case 491:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st492 }
		} else if data[p] >= 65 {
			goto st492
		}
	} else {
		goto st492
	}
	goto st0
st492:
	p++
	if p == pe { goto _test_eof492 }
	fallthrough
case 492:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st496
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st493 }
		} else if data[p] >= 65 {
			goto st493
		}
	} else {
		goto st493
	}
	goto st0
st493:
	p++
	if p == pe { goto _test_eof493 }
	fallthrough
case 493:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st496
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st494 }
		} else if data[p] >= 65 {
			goto st494
		}
	} else {
		goto st494
	}
	goto st0
st494:
	p++
	if p == pe { goto _test_eof494 }
	fallthrough
case 494:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st496
		case 125: goto tr323
	}
	if data[p] < 48 {
		if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st495 }
		} else if data[p] >= 65 {
			goto st495
		}
	} else {
		goto st495
	}
	goto st0
st495:
	p++
	if p == pe { goto _test_eof495 }
	fallthrough
case 495:
	switch data[p] {
		case 10: goto tr321
		case 32: goto tr322
		case 58: goto st496
		case 125: goto tr323
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr320 }
	goto st0
st496:
	p++
	if p == pe { goto _test_eof496 }
	fallthrough
case 496:
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st475 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st475 }
	} else {
		goto st475
	}
	goto st0
st497:
	p++
	if p == pe { goto _test_eof497 }
	fallthrough
case 497:
	if data[p] == 58 { goto st434 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st433 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st433 }
	} else {
		goto st433
	}
	goto st0
st498:
	p++
	if p == pe { goto _test_eof498 }
	fallthrough
case 498:
	if data[p] == 58 { goto st434 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st497 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st497 }
	} else {
		goto st497
	}
	goto st0
tr301:
// line 41 "parser.rl"
	{ mark = p }
	goto st499
st499:
	p++
	if p == pe { goto _test_eof499 }
	fallthrough
case 499:
// line 6407 "parser.go"
	switch data[p] {
		case 46: goto st251
		case 53: goto st500
		case 58: goto st434
	}
	if data[p] < 54 {
		if 48 <= data[p] && data[p] <= 52 { goto st431 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st498 }
		} else if data[p] >= 65 {
			goto st498
		}
	} else {
		goto st501
	}
	goto st0
st500:
	p++
	if p == pe { goto _test_eof500 }
	fallthrough
case 500:
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 54 {
		if 48 <= data[p] && data[p] <= 53 { goto st432 }
	} else if data[p] > 57 {
		if data[p] > 70 {
			if 97 <= data[p] && data[p] <= 102 { goto st497 }
		} else if data[p] >= 65 {
			goto st497
		}
	} else {
		goto st497
	}
	goto st0
st501:
	p++
	if p == pe { goto _test_eof501 }
	fallthrough
case 501:
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st497 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st497 }
	} else {
		goto st497
	}
	goto st0
tr302:
// line 41 "parser.rl"
	{ mark = p }
	goto st502
st502:
	p++
	if p == pe { goto _test_eof502 }
	fallthrough
case 502:
// line 6472 "parser.go"
	switch data[p] {
		case 46: goto st251
		case 58: goto st434
	}
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st501 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st498 }
	} else {
		goto st498
	}
	goto st0
tr303:
// line 41 "parser.rl"
	{ mark = p }
	goto st503
st503:
	p++
	if p == pe { goto _test_eof503 }
	fallthrough
case 503:
// line 6494 "parser.go"
	if data[p] == 58 { goto st491 }
	goto st0
tr304:
// line 41 "parser.rl"
	{ mark = p }
	goto st504
st504:
	p++
	if p == pe { goto _test_eof504 }
	fallthrough
case 504:
// line 6506 "parser.go"
	if data[p] == 58 { goto st434 }
	if data[p] < 65 {
		if 48 <= data[p] && data[p] <= 57 { goto st498 }
	} else if data[p] > 70 {
		if 97 <= data[p] && data[p] <= 102 { goto st498 }
	} else {
		goto st498
	}
	goto st0
st505:
	p++
	if p == pe { goto _test_eof505 }
	fallthrough
case 505:
	if data[p] == 108 { goto st506 }
	goto st0
st506:
	p++
	if p == pe { goto _test_eof506 }
	fallthrough
case 506:
	if data[p] == 97 { goto st507 }
	goto st0
st507:
	p++
	if p == pe { goto _test_eof507 }
	fallthrough
case 507:
	if data[p] == 115 { goto st508 }
	goto st0
st508:
	p++
	if p == pe { goto _test_eof508 }
	fallthrough
case 508:
	if data[p] == 115 { goto st509 }
	goto st0
st509:
	p++
	if p == pe { goto _test_eof509 }
	fallthrough
case 509:
	if data[p] == 32 { goto st510 }
	goto st0
st510:
	p++
	if p == pe { goto _test_eof510 }
	fallthrough
case 510:
	switch data[p] {
		case 32: goto st510
		case 61: goto st511
	}
	goto st0
st511:
	p++
	if p == pe { goto _test_eof511 }
	fallthrough
case 511:
	if data[p] == 32 { goto st512 }
	goto st0
st512:
	p++
	if p == pe { goto _test_eof512 }
	fallthrough
case 512:
	switch data[p] {
		case 32: goto st512
		case 99: goto tr575
		case 115: goto tr576
	}
	goto st0
tr575:
// line 41 "parser.rl"
	{ mark = p }
	goto st513
st513:
	p++
	if p == pe { goto _test_eof513 }
	fallthrough
case 513:
// line 6588 "parser.go"
	if data[p] == 108 { goto st514 }
	goto st0
st514:
	p++
	if p == pe { goto _test_eof514 }
	fallthrough
case 514:
	if data[p] == 105 { goto st515 }
	goto st0
st515:
	p++
	if p == pe { goto _test_eof515 }
	fallthrough
case 515:
	if data[p] == 101 { goto st516 }
	goto st0
st516:
	p++
	if p == pe { goto _test_eof516 }
	fallthrough
case 516:
	if data[p] == 110 { goto st517 }
	goto st0
st517:
	p++
	if p == pe { goto _test_eof517 }
	fallthrough
case 517:
	if data[p] == 116 { goto st518 }
	goto st0
st518:
	p++
	if p == pe { goto _test_eof518 }
	fallthrough
case 518:
	switch data[p] {
		case 10: goto tr448
		case 32: goto tr583
		case 125: goto tr584
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr582 }
	goto st0
tr576:
// line 41 "parser.rl"
	{ mark = p }
	goto st519
st519:
	p++
	if p == pe { goto _test_eof519 }
	fallthrough
case 519:
// line 6640 "parser.go"
	if data[p] == 101 { goto st520 }
	goto st0
st520:
	p++
	if p == pe { goto _test_eof520 }
	fallthrough
case 520:
	if data[p] == 114 { goto st521 }
	goto st0
st521:
	p++
	if p == pe { goto _test_eof521 }
	fallthrough
case 521:
	if data[p] == 118 { goto st522 }
	goto st0
st522:
	p++
	if p == pe { goto _test_eof522 }
	fallthrough
case 522:
	if data[p] == 101 { goto st523 }
	goto st0
st523:
	p++
	if p == pe { goto _test_eof523 }
	fallthrough
case 523:
	if data[p] == 114 { goto st518 }
	goto st0
st524:
	p++
	if p == pe { goto _test_eof524 }
	fallthrough
case 524:
	if data[p] == 115 { goto st525 }
	goto st0
st525:
	p++
	if p == pe { goto _test_eof525 }
	fallthrough
case 525:
	if data[p] == 108 { goto st526 }
	goto st0
st526:
	p++
	if p == pe { goto _test_eof526 }
	fallthrough
case 526:
	if data[p] == 32 { goto st527 }
	goto st0
st527:
	p++
	if p == pe { goto _test_eof527 }
	fallthrough
case 527:
	switch data[p] {
		case 32: goto st527
		case 61: goto st528
	}
	goto st0
st528:
	p++
	if p == pe { goto _test_eof528 }
	fallthrough
case 528:
	if data[p] == 32 { goto st529 }
	goto st0
st529:
	p++
	if p == pe { goto _test_eof529 }
	fallthrough
case 529:
	switch data[p] {
		case 32: goto st529
		case 102: goto tr595
		case 110: goto tr596
		case 111: goto tr597
		case 116: goto tr598
		case 121: goto tr599
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr594 }
	goto st0
tr594:
// line 41 "parser.rl"
	{ mark = p }
	goto st530
st530:
	p++
	if p == pe { goto _test_eof530 }
	fallthrough
case 530:
// line 6733 "parser.go"
	switch data[p] {
		case 10: goto tr465
		case 32: goto tr601
		case 125: goto tr602
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr600 }
	goto st0
tr595:
// line 41 "parser.rl"
	{ mark = p }
	goto st531
st531:
	p++
	if p == pe { goto _test_eof531 }
	fallthrough
case 531:
// line 6750 "parser.go"
	if data[p] == 97 { goto st532 }
	goto st0
st532:
	p++
	if p == pe { goto _test_eof532 }
	fallthrough
case 532:
	if data[p] == 108 { goto st533 }
	goto st0
st533:
	p++
	if p == pe { goto _test_eof533 }
	fallthrough
case 533:
	if data[p] == 115 { goto st534 }
	goto st0
st534:
	p++
	if p == pe { goto _test_eof534 }
	fallthrough
case 534:
	if data[p] == 101 { goto st530 }
	goto st0
tr596:
// line 41 "parser.rl"
	{ mark = p }
	goto st535
st535:
	p++
	if p == pe { goto _test_eof535 }
	fallthrough
case 535:
// line 6783 "parser.go"
	if data[p] == 111 { goto st530 }
	goto st0
tr597:
// line 41 "parser.rl"
	{ mark = p }
	goto st536
st536:
	p++
	if p == pe { goto _test_eof536 }
	fallthrough
case 536:
// line 6795 "parser.go"
	switch data[p] {
		case 102: goto st537
		case 110: goto st530
	}
	goto st0
st537:
	p++
	if p == pe { goto _test_eof537 }
	fallthrough
case 537:
	if data[p] == 102 { goto st530 }
	goto st0
tr598:
// line 41 "parser.rl"
	{ mark = p }
	goto st538
st538:
	p++
	if p == pe { goto _test_eof538 }
	fallthrough
case 538:
// line 6817 "parser.go"
	if data[p] == 114 { goto st539 }
	goto st0
st539:
	p++
	if p == pe { goto _test_eof539 }
	fallthrough
case 539:
	if data[p] == 117 { goto st534 }
	goto st0
tr599:
// line 41 "parser.rl"
	{ mark = p }
	goto st540
st540:
	p++
	if p == pe { goto _test_eof540 }
	fallthrough
case 540:
// line 6836 "parser.go"
	if data[p] == 101 { goto st541 }
	goto st0
st541:
	p++
	if p == pe { goto _test_eof541 }
	fallthrough
case 541:
	if data[p] == 115 { goto st530 }
	goto st0
st542:
	p++
	if p == pe { goto _test_eof542 }
	fallthrough
case 542:
	if data[p] == 105 { goto st543 }
	goto st0
st543:
	p++
	if p == pe { goto _test_eof543 }
	fallthrough
case 543:
	if data[p] == 112 { goto st544 }
	goto st0
st544:
	p++
	if p == pe { goto _test_eof544 }
	fallthrough
case 544:
	if data[p] == 32 { goto st545 }
	goto st0
st545:
	p++
	if p == pe { goto _test_eof545 }
	fallthrough
case 545:
	switch data[p] {
		case 32: goto st545
		case 61: goto st546
	}
	goto st0
st546:
	p++
	if p == pe { goto _test_eof546 }
	fallthrough
case 546:
	if data[p] == 32 { goto st547 }
	goto st0
st547:
	p++
	if p == pe { goto _test_eof547 }
	fallthrough
case 547:
	switch data[p] {
		case 32: goto st547
		case 102: goto tr616
		case 110: goto tr617
		case 111: goto tr618
		case 116: goto tr619
		case 121: goto tr620
	}
	if 48 <= data[p] && data[p] <= 49 { goto tr615 }
	goto st0
tr615:
// line 41 "parser.rl"
	{ mark = p }
	goto st548
st548:
	p++
	if p == pe { goto _test_eof548 }
	fallthrough
case 548:
// line 6908 "parser.go"
	switch data[p] {
		case 10: goto tr485
		case 32: goto tr622
		case 125: goto tr623
	}
	if 9 <= data[p] && data[p] <= 13 { goto tr621 }
	goto st0
tr616:
// line 41 "parser.rl"
	{ mark = p }
	goto st549
st549:
	p++
	if p == pe { goto _test_eof549 }
	fallthrough
case 549:
// line 6925 "parser.go"
	if data[p] == 97 { goto st550 }
	goto st0
st550:
	p++
	if p == pe { goto _test_eof550 }
	fallthrough
case 550:
	if data[p] == 108 { goto st551 }
	goto st0
st551:
	p++
	if p == pe { goto _test_eof551 }
	fallthrough
case 551:
	if data[p] == 115 { goto st552 }
	goto st0
st552:
	p++
	if p == pe { goto _test_eof552 }
	fallthrough
case 552:
	if data[p] == 101 { goto st548 }
	goto st0
tr617:
// line 41 "parser.rl"
	{ mark = p }
	goto st553
st553:
	p++
	if p == pe { goto _test_eof553 }
	fallthrough
case 553:
// line 6958 "parser.go"
	if data[p] == 111 { goto st548 }
	goto st0
tr618:
// line 41 "parser.rl"
	{ mark = p }
	goto st554
st554:
	p++
	if p == pe { goto _test_eof554 }
	fallthrough
case 554:
// line 6970 "parser.go"
	switch data[p] {
		case 102: goto st555
		case 110: goto st548
	}
	goto st0
st555:
	p++
	if p == pe { goto _test_eof555 }
	fallthrough
case 555:
	if data[p] == 102 { goto st548 }
	goto st0
tr619:
// line 41 "parser.rl"
	{ mark = p }
	goto st556
st556:
	p++
	if p == pe { goto _test_eof556 }
	fallthrough
case 556:
// line 6992 "parser.go"
	if data[p] == 114 { goto st557 }
	goto st0
st557:
	p++
	if p == pe { goto _test_eof557 }
	fallthrough
case 557:
	if data[p] == 117 { goto st552 }
	goto st0
tr620:
// line 41 "parser.rl"
	{ mark = p }
	goto st558
st558:
	p++
	if p == pe { goto _test_eof558 }
	fallthrough
case 558:
// line 7011 "parser.go"
	if data[p] == 101 { goto st559 }
	goto st0
st559:
	p++
	if p == pe { goto _test_eof559 }
	fallthrough
case 559:
	if data[p] == 115 { goto st548 }
	goto st0
	}
	_test_eof2: cs = 2; goto _test_eof; 
	_test_eof3: cs = 3; goto _test_eof; 
	_test_eof4: cs = 4; goto _test_eof; 
	_test_eof5: cs = 5; goto _test_eof; 
	_test_eof6: cs = 6; goto _test_eof; 
	_test_eof7: cs = 7; goto _test_eof; 
	_test_eof8: cs = 8; goto _test_eof; 
	_test_eof560: cs = 560; goto _test_eof; 
	_test_eof561: cs = 561; goto _test_eof; 
	_test_eof9: cs = 9; goto _test_eof; 
	_test_eof10: cs = 10; goto _test_eof; 
	_test_eof11: cs = 11; goto _test_eof; 
	_test_eof12: cs = 12; goto _test_eof; 
	_test_eof13: cs = 13; goto _test_eof; 
	_test_eof14: cs = 14; goto _test_eof; 
	_test_eof15: cs = 15; goto _test_eof; 
	_test_eof16: cs = 16; goto _test_eof; 
	_test_eof17: cs = 17; goto _test_eof; 
	_test_eof18: cs = 18; goto _test_eof; 
	_test_eof19: cs = 19; goto _test_eof; 
	_test_eof20: cs = 20; goto _test_eof; 
	_test_eof21: cs = 21; goto _test_eof; 
	_test_eof22: cs = 22; goto _test_eof; 
	_test_eof23: cs = 23; goto _test_eof; 
	_test_eof24: cs = 24; goto _test_eof; 
	_test_eof562: cs = 562; goto _test_eof; 
	_test_eof25: cs = 25; goto _test_eof; 
	_test_eof26: cs = 26; goto _test_eof; 
	_test_eof27: cs = 27; goto _test_eof; 
	_test_eof28: cs = 28; goto _test_eof; 
	_test_eof29: cs = 29; goto _test_eof; 
	_test_eof30: cs = 30; goto _test_eof; 
	_test_eof31: cs = 31; goto _test_eof; 
	_test_eof32: cs = 32; goto _test_eof; 
	_test_eof33: cs = 33; goto _test_eof; 
	_test_eof34: cs = 34; goto _test_eof; 
	_test_eof35: cs = 35; goto _test_eof; 
	_test_eof36: cs = 36; goto _test_eof; 
	_test_eof37: cs = 37; goto _test_eof; 
	_test_eof38: cs = 38; goto _test_eof; 
	_test_eof39: cs = 39; goto _test_eof; 
	_test_eof40: cs = 40; goto _test_eof; 
	_test_eof41: cs = 41; goto _test_eof; 
	_test_eof42: cs = 42; goto _test_eof; 
	_test_eof43: cs = 43; goto _test_eof; 
	_test_eof44: cs = 44; goto _test_eof; 
	_test_eof45: cs = 45; goto _test_eof; 
	_test_eof46: cs = 46; goto _test_eof; 
	_test_eof47: cs = 47; goto _test_eof; 
	_test_eof48: cs = 48; goto _test_eof; 
	_test_eof49: cs = 49; goto _test_eof; 
	_test_eof50: cs = 50; goto _test_eof; 
	_test_eof51: cs = 51; goto _test_eof; 
	_test_eof52: cs = 52; goto _test_eof; 
	_test_eof53: cs = 53; goto _test_eof; 
	_test_eof54: cs = 54; goto _test_eof; 
	_test_eof55: cs = 55; goto _test_eof; 
	_test_eof56: cs = 56; goto _test_eof; 
	_test_eof57: cs = 57; goto _test_eof; 
	_test_eof58: cs = 58; goto _test_eof; 
	_test_eof59: cs = 59; goto _test_eof; 
	_test_eof60: cs = 60; goto _test_eof; 
	_test_eof61: cs = 61; goto _test_eof; 
	_test_eof62: cs = 62; goto _test_eof; 
	_test_eof63: cs = 63; goto _test_eof; 
	_test_eof64: cs = 64; goto _test_eof; 
	_test_eof65: cs = 65; goto _test_eof; 
	_test_eof66: cs = 66; goto _test_eof; 
	_test_eof67: cs = 67; goto _test_eof; 
	_test_eof68: cs = 68; goto _test_eof; 
	_test_eof69: cs = 69; goto _test_eof; 
	_test_eof70: cs = 70; goto _test_eof; 
	_test_eof71: cs = 71; goto _test_eof; 
	_test_eof72: cs = 72; goto _test_eof; 
	_test_eof73: cs = 73; goto _test_eof; 
	_test_eof74: cs = 74; goto _test_eof; 
	_test_eof75: cs = 75; goto _test_eof; 
	_test_eof76: cs = 76; goto _test_eof; 
	_test_eof77: cs = 77; goto _test_eof; 
	_test_eof78: cs = 78; goto _test_eof; 
	_test_eof79: cs = 79; goto _test_eof; 
	_test_eof80: cs = 80; goto _test_eof; 
	_test_eof81: cs = 81; goto _test_eof; 
	_test_eof82: cs = 82; goto _test_eof; 
	_test_eof83: cs = 83; goto _test_eof; 
	_test_eof84: cs = 84; goto _test_eof; 
	_test_eof85: cs = 85; goto _test_eof; 
	_test_eof86: cs = 86; goto _test_eof; 
	_test_eof87: cs = 87; goto _test_eof; 
	_test_eof88: cs = 88; goto _test_eof; 
	_test_eof89: cs = 89; goto _test_eof; 
	_test_eof90: cs = 90; goto _test_eof; 
	_test_eof91: cs = 91; goto _test_eof; 
	_test_eof92: cs = 92; goto _test_eof; 
	_test_eof93: cs = 93; goto _test_eof; 
	_test_eof94: cs = 94; goto _test_eof; 
	_test_eof95: cs = 95; goto _test_eof; 
	_test_eof96: cs = 96; goto _test_eof; 
	_test_eof97: cs = 97; goto _test_eof; 
	_test_eof98: cs = 98; goto _test_eof; 
	_test_eof99: cs = 99; goto _test_eof; 
	_test_eof100: cs = 100; goto _test_eof; 
	_test_eof101: cs = 101; goto _test_eof; 
	_test_eof102: cs = 102; goto _test_eof; 
	_test_eof103: cs = 103; goto _test_eof; 
	_test_eof104: cs = 104; goto _test_eof; 
	_test_eof105: cs = 105; goto _test_eof; 
	_test_eof106: cs = 106; goto _test_eof; 
	_test_eof107: cs = 107; goto _test_eof; 
	_test_eof108: cs = 108; goto _test_eof; 
	_test_eof109: cs = 109; goto _test_eof; 
	_test_eof110: cs = 110; goto _test_eof; 
	_test_eof111: cs = 111; goto _test_eof; 
	_test_eof112: cs = 112; goto _test_eof; 
	_test_eof113: cs = 113; goto _test_eof; 
	_test_eof114: cs = 114; goto _test_eof; 
	_test_eof115: cs = 115; goto _test_eof; 
	_test_eof116: cs = 116; goto _test_eof; 
	_test_eof117: cs = 117; goto _test_eof; 
	_test_eof118: cs = 118; goto _test_eof; 
	_test_eof119: cs = 119; goto _test_eof; 
	_test_eof120: cs = 120; goto _test_eof; 
	_test_eof121: cs = 121; goto _test_eof; 
	_test_eof122: cs = 122; goto _test_eof; 
	_test_eof123: cs = 123; goto _test_eof; 
	_test_eof124: cs = 124; goto _test_eof; 
	_test_eof125: cs = 125; goto _test_eof; 
	_test_eof126: cs = 126; goto _test_eof; 
	_test_eof127: cs = 127; goto _test_eof; 
	_test_eof128: cs = 128; goto _test_eof; 
	_test_eof129: cs = 129; goto _test_eof; 
	_test_eof130: cs = 130; goto _test_eof; 
	_test_eof131: cs = 131; goto _test_eof; 
	_test_eof132: cs = 132; goto _test_eof; 
	_test_eof133: cs = 133; goto _test_eof; 
	_test_eof134: cs = 134; goto _test_eof; 
	_test_eof135: cs = 135; goto _test_eof; 
	_test_eof136: cs = 136; goto _test_eof; 
	_test_eof137: cs = 137; goto _test_eof; 
	_test_eof138: cs = 138; goto _test_eof; 
	_test_eof139: cs = 139; goto _test_eof; 
	_test_eof140: cs = 140; goto _test_eof; 
	_test_eof141: cs = 141; goto _test_eof; 
	_test_eof142: cs = 142; goto _test_eof; 
	_test_eof143: cs = 143; goto _test_eof; 
	_test_eof144: cs = 144; goto _test_eof; 
	_test_eof145: cs = 145; goto _test_eof; 
	_test_eof146: cs = 146; goto _test_eof; 
	_test_eof147: cs = 147; goto _test_eof; 
	_test_eof148: cs = 148; goto _test_eof; 
	_test_eof149: cs = 149; goto _test_eof; 
	_test_eof150: cs = 150; goto _test_eof; 
	_test_eof151: cs = 151; goto _test_eof; 
	_test_eof152: cs = 152; goto _test_eof; 
	_test_eof153: cs = 153; goto _test_eof; 
	_test_eof154: cs = 154; goto _test_eof; 
	_test_eof155: cs = 155; goto _test_eof; 
	_test_eof156: cs = 156; goto _test_eof; 
	_test_eof157: cs = 157; goto _test_eof; 
	_test_eof158: cs = 158; goto _test_eof; 
	_test_eof159: cs = 159; goto _test_eof; 
	_test_eof160: cs = 160; goto _test_eof; 
	_test_eof161: cs = 161; goto _test_eof; 
	_test_eof162: cs = 162; goto _test_eof; 
	_test_eof163: cs = 163; goto _test_eof; 
	_test_eof164: cs = 164; goto _test_eof; 
	_test_eof165: cs = 165; goto _test_eof; 
	_test_eof166: cs = 166; goto _test_eof; 
	_test_eof167: cs = 167; goto _test_eof; 
	_test_eof168: cs = 168; goto _test_eof; 
	_test_eof169: cs = 169; goto _test_eof; 
	_test_eof170: cs = 170; goto _test_eof; 
	_test_eof171: cs = 171; goto _test_eof; 
	_test_eof172: cs = 172; goto _test_eof; 
	_test_eof173: cs = 173; goto _test_eof; 
	_test_eof174: cs = 174; goto _test_eof; 
	_test_eof175: cs = 175; goto _test_eof; 
	_test_eof176: cs = 176; goto _test_eof; 
	_test_eof177: cs = 177; goto _test_eof; 
	_test_eof178: cs = 178; goto _test_eof; 
	_test_eof179: cs = 179; goto _test_eof; 
	_test_eof180: cs = 180; goto _test_eof; 
	_test_eof181: cs = 181; goto _test_eof; 
	_test_eof182: cs = 182; goto _test_eof; 
	_test_eof183: cs = 183; goto _test_eof; 
	_test_eof184: cs = 184; goto _test_eof; 
	_test_eof185: cs = 185; goto _test_eof; 
	_test_eof186: cs = 186; goto _test_eof; 
	_test_eof187: cs = 187; goto _test_eof; 
	_test_eof188: cs = 188; goto _test_eof; 
	_test_eof189: cs = 189; goto _test_eof; 
	_test_eof190: cs = 190; goto _test_eof; 
	_test_eof191: cs = 191; goto _test_eof; 
	_test_eof192: cs = 192; goto _test_eof; 
	_test_eof193: cs = 193; goto _test_eof; 
	_test_eof194: cs = 194; goto _test_eof; 
	_test_eof195: cs = 195; goto _test_eof; 
	_test_eof196: cs = 196; goto _test_eof; 
	_test_eof197: cs = 197; goto _test_eof; 
	_test_eof198: cs = 198; goto _test_eof; 
	_test_eof199: cs = 199; goto _test_eof; 
	_test_eof200: cs = 200; goto _test_eof; 
	_test_eof201: cs = 201; goto _test_eof; 
	_test_eof202: cs = 202; goto _test_eof; 
	_test_eof203: cs = 203; goto _test_eof; 
	_test_eof204: cs = 204; goto _test_eof; 
	_test_eof205: cs = 205; goto _test_eof; 
	_test_eof563: cs = 563; goto _test_eof; 
	_test_eof564: cs = 564; goto _test_eof; 
	_test_eof565: cs = 565; goto _test_eof; 
	_test_eof206: cs = 206; goto _test_eof; 
	_test_eof566: cs = 566; goto _test_eof; 
	_test_eof207: cs = 207; goto _test_eof; 
	_test_eof208: cs = 208; goto _test_eof; 
	_test_eof209: cs = 209; goto _test_eof; 
	_test_eof210: cs = 210; goto _test_eof; 
	_test_eof211: cs = 211; goto _test_eof; 
	_test_eof212: cs = 212; goto _test_eof; 
	_test_eof213: cs = 213; goto _test_eof; 
	_test_eof214: cs = 214; goto _test_eof; 
	_test_eof215: cs = 215; goto _test_eof; 
	_test_eof216: cs = 216; goto _test_eof; 
	_test_eof217: cs = 217; goto _test_eof; 
	_test_eof218: cs = 218; goto _test_eof; 
	_test_eof219: cs = 219; goto _test_eof; 
	_test_eof220: cs = 220; goto _test_eof; 
	_test_eof221: cs = 221; goto _test_eof; 
	_test_eof222: cs = 222; goto _test_eof; 
	_test_eof223: cs = 223; goto _test_eof; 
	_test_eof224: cs = 224; goto _test_eof; 
	_test_eof225: cs = 225; goto _test_eof; 
	_test_eof226: cs = 226; goto _test_eof; 
	_test_eof227: cs = 227; goto _test_eof; 
	_test_eof228: cs = 228; goto _test_eof; 
	_test_eof229: cs = 229; goto _test_eof; 
	_test_eof230: cs = 230; goto _test_eof; 
	_test_eof231: cs = 231; goto _test_eof; 
	_test_eof232: cs = 232; goto _test_eof; 
	_test_eof233: cs = 233; goto _test_eof; 
	_test_eof234: cs = 234; goto _test_eof; 
	_test_eof235: cs = 235; goto _test_eof; 
	_test_eof236: cs = 236; goto _test_eof; 
	_test_eof237: cs = 237; goto _test_eof; 
	_test_eof238: cs = 238; goto _test_eof; 
	_test_eof239: cs = 239; goto _test_eof; 
	_test_eof240: cs = 240; goto _test_eof; 
	_test_eof241: cs = 241; goto _test_eof; 
	_test_eof242: cs = 242; goto _test_eof; 
	_test_eof243: cs = 243; goto _test_eof; 
	_test_eof244: cs = 244; goto _test_eof; 
	_test_eof245: cs = 245; goto _test_eof; 
	_test_eof246: cs = 246; goto _test_eof; 
	_test_eof247: cs = 247; goto _test_eof; 
	_test_eof248: cs = 248; goto _test_eof; 
	_test_eof249: cs = 249; goto _test_eof; 
	_test_eof250: cs = 250; goto _test_eof; 
	_test_eof251: cs = 251; goto _test_eof; 
	_test_eof252: cs = 252; goto _test_eof; 
	_test_eof253: cs = 253; goto _test_eof; 
	_test_eof254: cs = 254; goto _test_eof; 
	_test_eof255: cs = 255; goto _test_eof; 
	_test_eof256: cs = 256; goto _test_eof; 
	_test_eof257: cs = 257; goto _test_eof; 
	_test_eof258: cs = 258; goto _test_eof; 
	_test_eof259: cs = 259; goto _test_eof; 
	_test_eof260: cs = 260; goto _test_eof; 
	_test_eof261: cs = 261; goto _test_eof; 
	_test_eof262: cs = 262; goto _test_eof; 
	_test_eof263: cs = 263; goto _test_eof; 
	_test_eof264: cs = 264; goto _test_eof; 
	_test_eof265: cs = 265; goto _test_eof; 
	_test_eof266: cs = 266; goto _test_eof; 
	_test_eof267: cs = 267; goto _test_eof; 
	_test_eof268: cs = 268; goto _test_eof; 
	_test_eof269: cs = 269; goto _test_eof; 
	_test_eof270: cs = 270; goto _test_eof; 
	_test_eof271: cs = 271; goto _test_eof; 
	_test_eof272: cs = 272; goto _test_eof; 
	_test_eof273: cs = 273; goto _test_eof; 
	_test_eof274: cs = 274; goto _test_eof; 
	_test_eof275: cs = 275; goto _test_eof; 
	_test_eof276: cs = 276; goto _test_eof; 
	_test_eof277: cs = 277; goto _test_eof; 
	_test_eof278: cs = 278; goto _test_eof; 
	_test_eof279: cs = 279; goto _test_eof; 
	_test_eof280: cs = 280; goto _test_eof; 
	_test_eof281: cs = 281; goto _test_eof; 
	_test_eof282: cs = 282; goto _test_eof; 
	_test_eof283: cs = 283; goto _test_eof; 
	_test_eof284: cs = 284; goto _test_eof; 
	_test_eof285: cs = 285; goto _test_eof; 
	_test_eof286: cs = 286; goto _test_eof; 
	_test_eof287: cs = 287; goto _test_eof; 
	_test_eof288: cs = 288; goto _test_eof; 
	_test_eof289: cs = 289; goto _test_eof; 
	_test_eof290: cs = 290; goto _test_eof; 
	_test_eof291: cs = 291; goto _test_eof; 
	_test_eof292: cs = 292; goto _test_eof; 
	_test_eof293: cs = 293; goto _test_eof; 
	_test_eof294: cs = 294; goto _test_eof; 
	_test_eof295: cs = 295; goto _test_eof; 
	_test_eof296: cs = 296; goto _test_eof; 
	_test_eof297: cs = 297; goto _test_eof; 
	_test_eof298: cs = 298; goto _test_eof; 
	_test_eof299: cs = 299; goto _test_eof; 
	_test_eof300: cs = 300; goto _test_eof; 
	_test_eof301: cs = 301; goto _test_eof; 
	_test_eof302: cs = 302; goto _test_eof; 
	_test_eof303: cs = 303; goto _test_eof; 
	_test_eof304: cs = 304; goto _test_eof; 
	_test_eof305: cs = 305; goto _test_eof; 
	_test_eof306: cs = 306; goto _test_eof; 
	_test_eof307: cs = 307; goto _test_eof; 
	_test_eof308: cs = 308; goto _test_eof; 
	_test_eof309: cs = 309; goto _test_eof; 
	_test_eof310: cs = 310; goto _test_eof; 
	_test_eof311: cs = 311; goto _test_eof; 
	_test_eof312: cs = 312; goto _test_eof; 
	_test_eof313: cs = 313; goto _test_eof; 
	_test_eof314: cs = 314; goto _test_eof; 
	_test_eof315: cs = 315; goto _test_eof; 
	_test_eof316: cs = 316; goto _test_eof; 
	_test_eof317: cs = 317; goto _test_eof; 
	_test_eof318: cs = 318; goto _test_eof; 
	_test_eof319: cs = 319; goto _test_eof; 
	_test_eof320: cs = 320; goto _test_eof; 
	_test_eof321: cs = 321; goto _test_eof; 
	_test_eof322: cs = 322; goto _test_eof; 
	_test_eof323: cs = 323; goto _test_eof; 
	_test_eof324: cs = 324; goto _test_eof; 
	_test_eof325: cs = 325; goto _test_eof; 
	_test_eof326: cs = 326; goto _test_eof; 
	_test_eof327: cs = 327; goto _test_eof; 
	_test_eof328: cs = 328; goto _test_eof; 
	_test_eof329: cs = 329; goto _test_eof; 
	_test_eof330: cs = 330; goto _test_eof; 
	_test_eof331: cs = 331; goto _test_eof; 
	_test_eof332: cs = 332; goto _test_eof; 
	_test_eof333: cs = 333; goto _test_eof; 
	_test_eof334: cs = 334; goto _test_eof; 
	_test_eof335: cs = 335; goto _test_eof; 
	_test_eof336: cs = 336; goto _test_eof; 
	_test_eof337: cs = 337; goto _test_eof; 
	_test_eof338: cs = 338; goto _test_eof; 
	_test_eof339: cs = 339; goto _test_eof; 
	_test_eof340: cs = 340; goto _test_eof; 
	_test_eof341: cs = 341; goto _test_eof; 
	_test_eof342: cs = 342; goto _test_eof; 
	_test_eof343: cs = 343; goto _test_eof; 
	_test_eof344: cs = 344; goto _test_eof; 
	_test_eof345: cs = 345; goto _test_eof; 
	_test_eof346: cs = 346; goto _test_eof; 
	_test_eof347: cs = 347; goto _test_eof; 
	_test_eof348: cs = 348; goto _test_eof; 
	_test_eof349: cs = 349; goto _test_eof; 
	_test_eof350: cs = 350; goto _test_eof; 
	_test_eof351: cs = 351; goto _test_eof; 
	_test_eof352: cs = 352; goto _test_eof; 
	_test_eof353: cs = 353; goto _test_eof; 
	_test_eof354: cs = 354; goto _test_eof; 
	_test_eof355: cs = 355; goto _test_eof; 
	_test_eof356: cs = 356; goto _test_eof; 
	_test_eof357: cs = 357; goto _test_eof; 
	_test_eof358: cs = 358; goto _test_eof; 
	_test_eof359: cs = 359; goto _test_eof; 
	_test_eof360: cs = 360; goto _test_eof; 
	_test_eof361: cs = 361; goto _test_eof; 
	_test_eof362: cs = 362; goto _test_eof; 
	_test_eof363: cs = 363; goto _test_eof; 
	_test_eof364: cs = 364; goto _test_eof; 
	_test_eof365: cs = 365; goto _test_eof; 
	_test_eof366: cs = 366; goto _test_eof; 
	_test_eof367: cs = 367; goto _test_eof; 
	_test_eof368: cs = 368; goto _test_eof; 
	_test_eof369: cs = 369; goto _test_eof; 
	_test_eof370: cs = 370; goto _test_eof; 
	_test_eof371: cs = 371; goto _test_eof; 
	_test_eof372: cs = 372; goto _test_eof; 
	_test_eof373: cs = 373; goto _test_eof; 
	_test_eof374: cs = 374; goto _test_eof; 
	_test_eof375: cs = 375; goto _test_eof; 
	_test_eof376: cs = 376; goto _test_eof; 
	_test_eof377: cs = 377; goto _test_eof; 
	_test_eof378: cs = 378; goto _test_eof; 
	_test_eof379: cs = 379; goto _test_eof; 
	_test_eof380: cs = 380; goto _test_eof; 
	_test_eof381: cs = 381; goto _test_eof; 
	_test_eof382: cs = 382; goto _test_eof; 
	_test_eof383: cs = 383; goto _test_eof; 
	_test_eof384: cs = 384; goto _test_eof; 
	_test_eof385: cs = 385; goto _test_eof; 
	_test_eof386: cs = 386; goto _test_eof; 
	_test_eof387: cs = 387; goto _test_eof; 
	_test_eof388: cs = 388; goto _test_eof; 
	_test_eof389: cs = 389; goto _test_eof; 
	_test_eof390: cs = 390; goto _test_eof; 
	_test_eof391: cs = 391; goto _test_eof; 
	_test_eof392: cs = 392; goto _test_eof; 
	_test_eof393: cs = 393; goto _test_eof; 
	_test_eof394: cs = 394; goto _test_eof; 
	_test_eof395: cs = 395; goto _test_eof; 
	_test_eof396: cs = 396; goto _test_eof; 
	_test_eof397: cs = 397; goto _test_eof; 
	_test_eof398: cs = 398; goto _test_eof; 
	_test_eof399: cs = 399; goto _test_eof; 
	_test_eof400: cs = 400; goto _test_eof; 
	_test_eof401: cs = 401; goto _test_eof; 
	_test_eof402: cs = 402; goto _test_eof; 
	_test_eof403: cs = 403; goto _test_eof; 
	_test_eof404: cs = 404; goto _test_eof; 
	_test_eof405: cs = 405; goto _test_eof; 
	_test_eof406: cs = 406; goto _test_eof; 
	_test_eof407: cs = 407; goto _test_eof; 
	_test_eof408: cs = 408; goto _test_eof; 
	_test_eof409: cs = 409; goto _test_eof; 
	_test_eof410: cs = 410; goto _test_eof; 
	_test_eof411: cs = 411; goto _test_eof; 
	_test_eof412: cs = 412; goto _test_eof; 
	_test_eof413: cs = 413; goto _test_eof; 
	_test_eof414: cs = 414; goto _test_eof; 
	_test_eof415: cs = 415; goto _test_eof; 
	_test_eof416: cs = 416; goto _test_eof; 
	_test_eof417: cs = 417; goto _test_eof; 
	_test_eof418: cs = 418; goto _test_eof; 
	_test_eof419: cs = 419; goto _test_eof; 
	_test_eof420: cs = 420; goto _test_eof; 
	_test_eof421: cs = 421; goto _test_eof; 
	_test_eof422: cs = 422; goto _test_eof; 
	_test_eof423: cs = 423; goto _test_eof; 
	_test_eof424: cs = 424; goto _test_eof; 
	_test_eof425: cs = 425; goto _test_eof; 
	_test_eof426: cs = 426; goto _test_eof; 
	_test_eof427: cs = 427; goto _test_eof; 
	_test_eof428: cs = 428; goto _test_eof; 
	_test_eof429: cs = 429; goto _test_eof; 
	_test_eof430: cs = 430; goto _test_eof; 
	_test_eof431: cs = 431; goto _test_eof; 
	_test_eof432: cs = 432; goto _test_eof; 
	_test_eof433: cs = 433; goto _test_eof; 
	_test_eof434: cs = 434; goto _test_eof; 
	_test_eof435: cs = 435; goto _test_eof; 
	_test_eof436: cs = 436; goto _test_eof; 
	_test_eof437: cs = 437; goto _test_eof; 
	_test_eof438: cs = 438; goto _test_eof; 
	_test_eof439: cs = 439; goto _test_eof; 
	_test_eof440: cs = 440; goto _test_eof; 
	_test_eof441: cs = 441; goto _test_eof; 
	_test_eof442: cs = 442; goto _test_eof; 
	_test_eof443: cs = 443; goto _test_eof; 
	_test_eof444: cs = 444; goto _test_eof; 
	_test_eof445: cs = 445; goto _test_eof; 
	_test_eof446: cs = 446; goto _test_eof; 
	_test_eof447: cs = 447; goto _test_eof; 
	_test_eof448: cs = 448; goto _test_eof; 
	_test_eof449: cs = 449; goto _test_eof; 
	_test_eof450: cs = 450; goto _test_eof; 
	_test_eof451: cs = 451; goto _test_eof; 
	_test_eof452: cs = 452; goto _test_eof; 
	_test_eof453: cs = 453; goto _test_eof; 
	_test_eof454: cs = 454; goto _test_eof; 
	_test_eof455: cs = 455; goto _test_eof; 
	_test_eof456: cs = 456; goto _test_eof; 
	_test_eof457: cs = 457; goto _test_eof; 
	_test_eof458: cs = 458; goto _test_eof; 
	_test_eof459: cs = 459; goto _test_eof; 
	_test_eof460: cs = 460; goto _test_eof; 
	_test_eof461: cs = 461; goto _test_eof; 
	_test_eof462: cs = 462; goto _test_eof; 
	_test_eof463: cs = 463; goto _test_eof; 
	_test_eof464: cs = 464; goto _test_eof; 
	_test_eof465: cs = 465; goto _test_eof; 
	_test_eof466: cs = 466; goto _test_eof; 
	_test_eof467: cs = 467; goto _test_eof; 
	_test_eof468: cs = 468; goto _test_eof; 
	_test_eof469: cs = 469; goto _test_eof; 
	_test_eof470: cs = 470; goto _test_eof; 
	_test_eof471: cs = 471; goto _test_eof; 
	_test_eof472: cs = 472; goto _test_eof; 
	_test_eof473: cs = 473; goto _test_eof; 
	_test_eof474: cs = 474; goto _test_eof; 
	_test_eof475: cs = 475; goto _test_eof; 
	_test_eof476: cs = 476; goto _test_eof; 
	_test_eof477: cs = 477; goto _test_eof; 
	_test_eof478: cs = 478; goto _test_eof; 
	_test_eof479: cs = 479; goto _test_eof; 
	_test_eof480: cs = 480; goto _test_eof; 
	_test_eof481: cs = 481; goto _test_eof; 
	_test_eof482: cs = 482; goto _test_eof; 
	_test_eof483: cs = 483; goto _test_eof; 
	_test_eof484: cs = 484; goto _test_eof; 
	_test_eof485: cs = 485; goto _test_eof; 
	_test_eof486: cs = 486; goto _test_eof; 
	_test_eof487: cs = 487; goto _test_eof; 
	_test_eof488: cs = 488; goto _test_eof; 
	_test_eof489: cs = 489; goto _test_eof; 
	_test_eof490: cs = 490; goto _test_eof; 
	_test_eof491: cs = 491; goto _test_eof; 
	_test_eof492: cs = 492; goto _test_eof; 
	_test_eof493: cs = 493; goto _test_eof; 
	_test_eof494: cs = 494; goto _test_eof; 
	_test_eof495: cs = 495; goto _test_eof; 
	_test_eof496: cs = 496; goto _test_eof; 
	_test_eof497: cs = 497; goto _test_eof; 
	_test_eof498: cs = 498; goto _test_eof; 
	_test_eof499: cs = 499; goto _test_eof; 
	_test_eof500: cs = 500; goto _test_eof; 
	_test_eof501: cs = 501; goto _test_eof; 
	_test_eof502: cs = 502; goto _test_eof; 
	_test_eof503: cs = 503; goto _test_eof; 
	_test_eof504: cs = 504; goto _test_eof; 
	_test_eof505: cs = 505; goto _test_eof; 
	_test_eof506: cs = 506; goto _test_eof; 
	_test_eof507: cs = 507; goto _test_eof; 
	_test_eof508: cs = 508; goto _test_eof; 
	_test_eof509: cs = 509; goto _test_eof; 
	_test_eof510: cs = 510; goto _test_eof; 
	_test_eof511: cs = 511; goto _test_eof; 
	_test_eof512: cs = 512; goto _test_eof; 
	_test_eof513: cs = 513; goto _test_eof; 
	_test_eof514: cs = 514; goto _test_eof; 
	_test_eof515: cs = 515; goto _test_eof; 
	_test_eof516: cs = 516; goto _test_eof; 
	_test_eof517: cs = 517; goto _test_eof; 
	_test_eof518: cs = 518; goto _test_eof; 
	_test_eof519: cs = 519; goto _test_eof; 
	_test_eof520: cs = 520; goto _test_eof; 
	_test_eof521: cs = 521; goto _test_eof; 
	_test_eof522: cs = 522; goto _test_eof; 
	_test_eof523: cs = 523; goto _test_eof; 
	_test_eof524: cs = 524; goto _test_eof; 
	_test_eof525: cs = 525; goto _test_eof; 
	_test_eof526: cs = 526; goto _test_eof; 
	_test_eof527: cs = 527; goto _test_eof; 
	_test_eof528: cs = 528; goto _test_eof; 
	_test_eof529: cs = 529; goto _test_eof; 
	_test_eof530: cs = 530; goto _test_eof; 
	_test_eof531: cs = 531; goto _test_eof; 
	_test_eof532: cs = 532; goto _test_eof; 
	_test_eof533: cs = 533; goto _test_eof; 
	_test_eof534: cs = 534; goto _test_eof; 
	_test_eof535: cs = 535; goto _test_eof; 
	_test_eof536: cs = 536; goto _test_eof; 
	_test_eof537: cs = 537; goto _test_eof; 
	_test_eof538: cs = 538; goto _test_eof; 
	_test_eof539: cs = 539; goto _test_eof; 
	_test_eof540: cs = 540; goto _test_eof; 
	_test_eof541: cs = 541; goto _test_eof; 
	_test_eof542: cs = 542; goto _test_eof; 
	_test_eof543: cs = 543; goto _test_eof; 
	_test_eof544: cs = 544; goto _test_eof; 
	_test_eof545: cs = 545; goto _test_eof; 
	_test_eof546: cs = 546; goto _test_eof; 
	_test_eof547: cs = 547; goto _test_eof; 
	_test_eof548: cs = 548; goto _test_eof; 
	_test_eof549: cs = 549; goto _test_eof; 
	_test_eof550: cs = 550; goto _test_eof; 
	_test_eof551: cs = 551; goto _test_eof; 
	_test_eof552: cs = 552; goto _test_eof; 
	_test_eof553: cs = 553; goto _test_eof; 
	_test_eof554: cs = 554; goto _test_eof; 
	_test_eof555: cs = 555; goto _test_eof; 
	_test_eof556: cs = 556; goto _test_eof; 
	_test_eof557: cs = 557; goto _test_eof; 
	_test_eof558: cs = 558; goto _test_eof; 
	_test_eof559: cs = 559; goto _test_eof; 

	_test_eof: {}
	if p == eof {
	switch cs {
	case 560:
// line 78 "parser.rl"
	{
		port := cur.(*cPort)
		conf.Ports[port.Port] = port
		cur = nil
	}
	break
	case 562, 563:
// line 134 "parser.rl"
	{
		oper := cur.(*cOper)
		conf.Opers[oper.Username] = oper
		cur = nil
	}
	break
// line 7607 "parser.go"
	}
	}

	_out: {}
	}

// line 228 "parser.rl"

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

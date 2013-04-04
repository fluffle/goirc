package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// DISCLAIMER: will probably not fix everything. Notable omissions:
//   - conn.State will not be removed, if you were using it.
//   - conn.ER/ED will not be removed; exposing them was a Bad Idea, and
//     the required changes are quite a challenge to express programatically.

func init() {
	register(goircFix)
}

var goircFix = fix{
	"goirc",
	"2013-03-24",
	goircNewApi,
	`Update code that uses goirc/client to new API.`,
}

var goircConstants = map[string]string{
	`"REGISTER"`: "REGISTER",
	`"CONNECTED"`: "CONNECTED",
	`"DISCONNECTED"`: "DISCONNECTED",
	`"ACTION"`: "ACTION",
	`"AWAY"`: "AWAY",
	`"CTCP"`: "CTCP",
	`"CTCPREPLY"`: "CTCPREPLY",
	`"INVITE"`: "INVITE",
	`"JOIN"`: "JOIN",
	`"KICK"`: "KICK",
	`"MODE"`: "MODE",
	`"NICK"`: "NICK",
	`"NOTICE"`: "NOTICE",
	`"OPER"`: "OPER",
	`"PART"`: "PART",
	`"PASS"`: "PASS",
	`"PING"`: "PING",
	`"PONG"`: "PONG",
	`"PRIVMSG"`: "PRIVMSG",
	`"QUIT"`: "QUIT",
	`"TOPIC"`: "TOPIC",
	`"USER"`: "USER",
	`"VERSION"`: "VERSION",
	`"VHOST"`: "VHOST",
	`"WHO"`: "WHO",
	`"WHOIS"`: "WHOIS",
}

var goircStructToConfig = map[string]string{
	"Host":      "Server",
	"Network":   "Server",
	"NewNick":   "NewNick",
	"SSL":	     "SSL",
	"SSLConfig": "SSLConfig",
	"PingFreq":  "PingFreq",
	"Flood":     "Flood",
}

var goircStructToMethod = map[string]string{
	"Me":        "Me",
	"ST":        "StateTracker",
	"Connected": "Connected",
}

var goircMethodRename = map[string]string{
	"AddHandler": "HandleFunc",
	"Connect":    "ConnectTo",
}

func addCall(t ast.Expr, method string) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   t,
			Sel: ast.NewIdent(method),
		},
	}
}

func goircNewApi(f *ast.File) bool {
	state := stateApi(f)
	client := clientApi(f)
	return client || state
}

func stateApi(f *ast.File) bool {
	spec := importSpec(f, "github.com/fluffle/goirc/state")
	if spec == nil {
		return false
	}
	stateImport := "state"
	if spec.Name != nil {
		stateImport = spec.Name.Name
	}
	return renameFixTab(f, []rename{
		{"github.com/fluffle/goirc/state", "",
			stateImport + ".StateTracker", stateImport + ".Tracker"},
		{"github.com/fluffle/goirc/state", "",
			stateImport + ".MockStateTracker", stateImport + ".MockTracker"},
	})
}

func clientApi(f *ast.File) bool {
	spec := importSpec(f, "github.com/fluffle/goirc/client")
	if spec == nil {
		return false
	}
	clientImport := "client"
	if spec.Name != nil {
		clientImport = spec.Name.Name
	}
	fixed := false

	maybeReplaceBasicLit := func (expr *ast.Expr) {
		str, ok := (*expr).(*ast.BasicLit)
		if !ok || str == nil || str.Kind != token.STRING { return }
		if repl, ok := goircConstants[strings.ToUpper(str.Value)]; ok {
			*expr = &ast.SelectorExpr{
				ast.NewIdent(clientImport),
				ast.NewIdent(repl),
			}
			fixed = true
		}
	}

	maybeReplaceConnSelectors := func (expr *ast.Expr) {
		sel, ok := (*expr).(*ast.SelectorExpr)
		if !ok || !isClientConn(sel.X, clientImport) { return }
		name := sel.Sel.String()
		if rep, ok := goircStructToConfig[name]; ok {
			sel.X   = addCall(sel.X, "Config")
			sel.Sel = ast.NewIdent(rep)
			fixed = true
		} else if meth, ok := goircStructToMethod[name]; ok {
			*expr = addCall(sel.X, meth)
			fixed = true
		} else if meth, ok := goircMethodRename[name]; ok {
			sel.Sel = ast.NewIdent(meth)
			fixed = true
		}
	}

	walk(f, func(n interface{}) {
		if expr, ok := n.(*ast.Expr); ok {
			maybeReplaceBasicLit(expr)
			maybeReplaceConnSelectors(expr)
		}
		if expr, ok := n.(*ast.CallExpr); ok {
			if sel, ok := n.(*ast.SelectorExpr); ok &&
				isPkgDot(sel, clientImport, "Client") {
				// s/Client/SimpleClient/
				sel.Sel = ast.NewIdent("SimpleClient")
				// and delete the last arg from args
				expr.Args = expr.Args[:3]
				fixed = true
			}
		}
	})
	return fixed
}

func isClientConn(t ast.Expr, pkg string) bool {
	id, ok := t.(*ast.Ident)
	if !ok || id.Obj == nil { return false }
	switch dec := id.Obj.Decl.(type) {
	case *ast.ValueSpec:
		// Declared with var X Type
		return dec.Type != nil && isPtrPkgDot(dec.Type, pkg, "Conn")
	case *ast.AssignStmt:
		// Declared with X := Expr producing Type
		// NOTE: not taking care of multiple-assignment case atm!
		switch rhs := dec.Rhs[0].(type) {
		case *ast.CallExpr:
			// X := client.Client() or client.SimpleClient()
			return isPkgDot(rhs.Fun, pkg, "Client") ||
				isPkgDot(rhs.Fun, pkg, "SimpleClient")
		case *ast.UnaryExpr:
			// X := &client.Conn{}
			lit, ok := rhs.X.(*ast.CompositeLit)
			return ok && isPkgDot(lit.Type, pkg, "Conn")
		case *ast.CompositeLit:
			// X := client.Conn{}
			return isPkgDot(lit.Type, pkg, "Conn")
		default:
			fmt.Printf("rhs: %#v\n", rhs)
		}
	case *ast.Field:
		// Declared with func f(X Type)
		return isPkgDot(dec.Type, pkg, "Conn")
	default:
		fmt.Printf("dec: %#v\n", dec)
	}
	return false
}

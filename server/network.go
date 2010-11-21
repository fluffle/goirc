package server

type Network struct {
	nodes map[string]*Node
	chans map[string]*Channel
	nicks map[string]*Nick
	tree  *NetMap
}


package server

type NetMap struct {
	// To get to a specific Node, go via..
	via map[*Node]*Node

	// Network links, from server's perspective
	links *Link
}

type Link struct {
	node *Node
	hops int
	links []*Link
}

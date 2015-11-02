// Package client implements an IRC client. It handles protocol basics
// such as initial connection and responding to server PINGs, and has
// optional state tracking support which will keep tabs on every nick
// present in the same channels as the client. Other features include
// SSL support, automatic splitting of long lines, and panic recovery
// for handlers.
//
// Incoming IRC messages are parsed into client.Line structs and trigger
// events based on the IRC verb (e.g. PRIVMSG) of the message. Handlers
// for these events conform to the client.Handler interface; a HandlerFunc
// type to wrap bare functions is provided a-la the net/http package.
//
// Creating a client, adding a handler and connecting to a server looks
// soemthing like this, for the simple case:
//
//     // Create a new client, which will connect with the nick "myNick"
//     irc := client.SimpleClient("myNick")
//
//     // Add a handler that waits for the "disconnected" event and
//     // closes a channel to signal everything is done.
//     disconnected := make(chan struct{})
//     c.HandleFunc("disconnected", func(c *client.Conn, l *client.Line) {
//         close(disconnected)
//     })
//
//     // Connect to an IRC server.
//     if err := c.ConnectTo("irc.freenode.net"); err != nil {
//         log.Fatalf("Connection error: %v\n", err)
//     }
//
//     // Wait for disconnection.
//     <-disconnected
//
package client

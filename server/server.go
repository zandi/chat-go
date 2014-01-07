/* chat-server.go

simple implementation of a chat server.
*/

package main

import (
	"fmt"
	"net"
	"os"
	"github.com/zandi/chat-go"
)

/* called as a goroutine to handle each
individual connection.

For now, just a simple echo server
*/
func handleConnection(client net.Conn, u chat.User, newmsg chan<- string) {
	ch_recv := make(chan chat.Message)
	ch_send := make(chan chat.Message)

	go chat.MessageReceiver(client, ch_recv)
	go chat.MessageSender(client, ch_send)

	for {
		select {
		case m, ok := <-ch_recv:
			if !ok {
				//client disconnected, close sender helper and quit
				close(ch_send)
				return
			} else {
				//new message from client
				newmsg<- u.Name
				u.In <- m
			}
		case m := <-u.Out:
			//new message to client
			ch_send<- m
		}
	}
}

func identify(c net.Conn) (chat.User, error) {
			buf := make([]byte, 1024)
			n, err := c.Read(buf)
			if err != nil {
				return chat.User{"",nil,nil}, err
			}

			username := string(buf[:n])
			in := make(chan chat.Message)
			out := make(chan chat.Message)
			u := chat.User{username, in, out}
			return u, nil
}

func main() {
	fmt.Println("Server starting...")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error: ",err)
		os.Exit(0)
	}
	fmt.Println("Listening on ", ln.Addr())

	r := new(chat.Router)
	newuser := make(chan chat.User)
	newmsg := make(chan string)
	go r.Route(newuser,newmsg)

	for {
		conn, err := ln.Accept()

		if err != nil {
			//some sort of error
			fmt.Println("Error: ",err)
		} else {
			u, err := identify(conn)
			if err != nil {
				fmt.Println(conn.RemoteAddr(),"Error identifying user")
				conn.Close()
				continue
			}
			newuser <- u
			fmt.Println(conn.RemoteAddr()," has joined as ",u.Name)
			go handleConnection(conn, u, newmsg)
		}
	}
}


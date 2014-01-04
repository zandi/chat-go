/* chat.go

Serves as a file to hold common structs and types

*/

package chat

import (
	"fmt"
	"net"
	"time"
	"encoding/json"
)

type Message struct {
	Source string
	Dest string //will implement once we start routing
	Text string
}

type Router struct {
	In_chans map[string](<-chan Message)
	Out_chans map[string](chan<- Message)
}

/*
only meant for use from the router's perspective 
*/
type User struct {
	Name string
	In chan Message
	Out chan Message
}

/*
init the maps in our router
*/
func (r *Router) init() {
	r.In_chans = make(map[string](<-chan Message))
	r.Out_chans = make(map[string](chan<- Message))
}

/*
actual function where routing logic is kept

We should really use buffered chans so the core
message router doesn't have to block each and
every time (?)

newuser delivers new users to the router
newmsg carries the name of the source of a waiting message
*/
func (r *Router) Route(newuser chan User, newmsg chan string) {
	r.init()
	for {
		select {
			case u := <-newuser:
				r.AddUser(u)
			case src := <-newmsg:
				m := <-r.In_chans[src]
				r.Out_chans[m.Dest]<- m
				fmt.Println("msg: ",src," -> ",m.Dest,": ",m)
		}

		//just so we don't hog cpu doing nothing
		time.Sleep(2000 * time.Millisecond)
	}
}

/*
add a user to the router

Users which have left are automatically removed
if their channel is closed. Thus, the connection
handler must close the channel
*/
func (r *Router) AddUser(u User) {
	r.In_chans[u.Name] = u.In
	r.Out_chans[u.Name] = u.Out
}

/*
remove a user. This is done automatically, so shouldn't
need to be called by outside functions
*/
func (r *Router) removeUser(u User) {
	delete(r.In_chans, u.Name)
	delete(r.Out_chans, u.Name)
}

/*
meant to be run as a goroutine. Sends all messages
received over the channel ch to the connected 
client c

returns only once ch is closed (no more messages to send)
*/
func MessageSender(c net.Conn, ch chan Message) {
	for {
		m, ok := <-ch
		if !ok {
			//channel closed, so client has left.
			//receiver will have closed connection already
			return
		}
		err := WriteMessage(c, m)
		if err != nil {
			fmt.Println(c.RemoteAddr()," WriteMessage: ",err)
		}
	}
}

/*
meant to be run as a goroutine. Receives messages
from connected client c and sends them to channel
ch

returns only once the client has disconnected, and
signals this to the user by closing the channel ch
*/
func MessageReceiver(c net.Conn, ch chan Message) {
	for {
		m, err := ReadMessage(c)
		if err != nil {
			fmt.Println(c.RemoteAddr()," ReadMessage: ",err)
			if err.Error() == "EOF" {
				c.Close()
				close(ch)
				return
			}
		} else {
			ch<- *m
		}
	}
}
/*
reads a message from Connection c, writing
to m, returning any errors. The message is
assumed to have been sent in JSON format

Errors from both conn.Read and json.Unmarshal
are placed in return value, so we may possible lose
some information here..
*/
func ReadMessage(c net.Conn) (*Message, error) {
	buf := make([]byte, 1024) //using JSON, how large must this be?
	var n int
	var err error
	if n, err = c.Read(buf); err != nil {
		if err.Error() == "EOF" {
			return nil, err
		}
	}

	//note that it's possible for Unmarshal to encounter
	//an error, but successfully unmarshal the json
	var m Message
	err = json.Unmarshal(buf[:n], &m)
	return &m, err
}

/*
writes the message m to connection c, returning
any errors. The message will me written in
JSON format
*/
func WriteMessage(c net.Conn, m Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if _, err := c.Write(b); err != nil {
		return err
	}
	return nil
}


/* chat-client.go

implements a simple client.
we take server as an argument, then
prompt the user for their username

*/

package main

import (
	"fmt"
	"bufio"
	"net"
	"os"
	"strings"
	"github.com/zandi/chat-go/chat"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments.")
		fmt.Println("USAGE: client [server]")
		fmt.Println("eg: 'client localhost:8080'")
		os.Exit(1)
	}

	username := ""
	for len(username) == 0 || len(username) > 255 {
		fmt.Print("Username: ")
		fmt.Scan(&username)
	}

	server, err := net.Dial("tcp", os.Args[1])
	if  err != nil {
		fmt.Println("Error: ",err)
		os.Exit(1)
	}
	fmt.Println("Connected to ",server.RemoteAddr())

	//identify to server. improve this
	fmt.Println("Identifying...")
	buf := []byte(username)
	_, err = server.Write(buf)
	if err != nil {
		fmt.Println("Error identifying to server: ",err)
		server.Close()
		os.Exit(1)
	}

	fmt.Println("exit by typing '/exit', or sending EOF (Ctrl+D)")

	var msg chat.Message
	msg.Source = username
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		textbuf, err := r.ReadString('\n')
		if err != nil {
			fmt.Println("ReadString: ",err)
			os.Exit(1)
		}

		textbuf = strings.TrimRight(textbuf, "\n")
		if textbuf == "/exit" {
			server.Close()
			os.Exit(0)
		}

		//todo: if no destination, broadcast
		//for now: if no destination, echo
		textslice := strings.Split(textbuf,":")
		if len(textslice) == 0 {
			continue
		} else if len(textslice) == 1 {
			msg.Dest = msg.Source
			msg.Text = textslice[0]
		} else {
			msg.Dest = textslice[0]
			msg.Text = strings.Join(textslice[1:],"")
		}

		err = chat.WriteMessage(server, msg)
		if err != nil {
			fmt.Println("WriteMessage: ",err)
		}

		msg, err := chat.ReadMessage(server)
		if err != nil {
			fmt.Println("ReadMessage: ",err)
		} else {
			fmt.Println(msg.Source,": ",msg.Text)
		}
	}

	server.Close()
}

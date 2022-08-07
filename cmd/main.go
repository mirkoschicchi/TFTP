package main

import (
	"fmt"
	"net"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/client"
	"github.com/mirkoschicchi/TFTP/internal/app/server"
)

func main() {

	role := os.Args[1]

	if role == "server" {
		fmt.Println("Starting the server and listening for incoming connections")
		var server server.Server
		err := server.Listen()
		if err != nil {
			fmt.Printf("The server has failed during listening: %+v", err)
		}
	} else {
		fmt.Println("Starting the client and connecting to the server")
		var client client.Client = client.NewClient()

		remoteAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:69")
		if err != nil {
			fmt.Printf("Error encountered while resolving remote addr: %+v", err)
		}
		err = client.Connect(remoteAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot connect to server: %+v", err)
		}
	}

}

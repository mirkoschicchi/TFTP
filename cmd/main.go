package main

import (
	"log"
	"net"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/client"
	"github.com/mirkoschicchi/TFTP/internal/app/logger"
	"github.com/mirkoschicchi/TFTP/internal/app/server"
)

func main() {
	role := os.Args[1]

	if role == "server" {
		logger.Info("Starting the server and listening for incoming connections")
		s := server.NewServer()
		err := s.Listen()
		if err != nil {
			logger.Fatal("The server has failed during listening: %+v", err)
		}
	} else {
		logger.Info("Starting the client and connecting to the server")
		var client client.Client = client.NewClient()

		remoteAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:69")
		if err != nil {
			logger.Fatal("Error encountered while resolving remote addr: %+v", err)
		}
		err = client.RequestFile(remoteAddr, os.Args[2])
		if err != nil {
			log.Fatalf("Cannot connect to server: %+v", err)
		}
	}
}

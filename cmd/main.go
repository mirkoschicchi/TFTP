package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/client"
	"github.com/mirkoschicchi/TFTP/internal/app/logger"
	"github.com/mirkoschicchi/TFTP/internal/app/server"
)

var (
	isServer      *bool
	isClient      *bool
	remoteAddress *string
)

func init() {
	isServer = flag.Bool("server", false, "Set this to true to spawn a TFTP server")
	isClient = flag.Bool("client", false, "Set this to true to spawn a TFTP client")
	remoteAddress = flag.String("remote", "127.0.0.1:69", "The address of the TFTP server")
}

func main() {
	flag.Parse()
	if (!*isServer && !*isClient) || (*isServer && *isClient) {
		panic("You have to specify if you want to run a server or a client!")
	}

	if *isServer {
		logger.Info("Starting the server and listening for incoming connections")
		s := server.NewServer()
		err := s.Listen()
		if err != nil {
			logger.Fatal("The server has failed during listening: %+v", err)
		}
	} else if *isClient {
		logger.Info("Starting the client and connecting to the server")
		var client client.Client = client.NewClient()

		remoteAddr, err := net.ResolveUDPAddr("udp4", *remoteAddress)
		if err != nil {
			logger.Fatal("Error encountered while resolving remote addr: %+v", err)
		}
		err = client.RequestFile(remoteAddr, os.Args[3])
		if err != nil {
			log.Fatalf("Cannot connect to server: %+v", err)
		}
	}
}

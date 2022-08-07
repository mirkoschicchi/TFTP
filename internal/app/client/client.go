package client

import (
	"fmt"
	"net"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/packets"
	"github.com/mirkoschicchi/TFTP/internal/app/utils"
	"github.com/pkg/errors"
)

type Client struct {
	TID  int
	Conn *net.UDPConn
}

func NewClient() Client {
	clientTID := utils.GetRandomTID()

	return Client{TID: clientTID}
}

func (c *Client) Connect(serverAddr *net.UDPAddr) error {
	var localAddress *net.UDPAddr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: c.TID}

	fmt.Printf("The client local address is %+v\n", localAddress)

	initialConnection, err := net.DialUDP("udp4", localAddress, serverAddr)
	if err != nil {
		return errors.Wrap(err, "cannot connect to server")
	}
	c.Conn = initialConnection

	fmt.Printf("Connection to server at %+v has been created\n", serverAddr)

	rrqPacket := packets.NewRRQPacket("/Users/mirko/Desktop/TFTP/testdata/hello.txt", packets.Netascii)
	_, err = c.Conn.Write(rrqPacket.Bytes())
	if err != nil {
		return errors.Wrap(err, "cannot write to server")
	}

	fmt.Println("I have sent the first RRQ packet to the server")
	initialConnection.Close()

	newConnection, err := net.ListenUDP("udp4", localAddress)
	if err != nil {
		return errors.Wrap(err, "error while listening for incoming UDP connections")
	}
	defer newConnection.Close()

	fmt.Print("New connection has been created\n")
	var receivedBytes []byte
	var isFinalBlock bool = false

	for !isFinalBlock {
		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		_, remoteAddr, err := newConnection.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		parsedPacket, err := packets.ParsePacket(buf)
		if err != nil {
			return errors.Wrap(err, "cannot parse incoming packet")
		}
		// fmt.Printf(">>> I have received the following packet data: %+v\n", parsedPacket)

		receivedBytes = append(receivedBytes, parsedPacket.(packets.DataPacket).Data...)
		if len(parsedPacket.(packets.DataPacket).Data) < utils.MAX_DATA_FIELD_LENGTH {
			isFinalBlock = true
		}

		ackPacket := packets.NewAckPacket(parsedPacket.(packets.DataPacket).BlockNumber)

		_, err = newConnection.WriteToUDP(ackPacket.Bytes(), remoteAddr)
		if err != nil {
			return errors.Wrap(err, "cannot write to server")
		}
		// fmt.Printf(">>> I have sent the following ack packet: %+v\n", ackPacket)
	}

	f, err := os.Create("received1.txt")
	if err != nil {
		return errors.Wrap(err, "cannot create file to be received")
	}
	f.WriteString(string(receivedBytes))

	return nil
}

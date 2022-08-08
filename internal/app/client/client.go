package client

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/mirkoschicchi/TFTP/internal/app/logger"
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

func (c *Client) RequestFile(serverAddr *net.UDPAddr, requestedFilePath string) error {
	var localAddress *net.UDPAddr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: c.TID}

	fmt.Printf("The client local address is %+v\n", localAddress)

	initialConnection, err := net.DialUDP("udp4", localAddress, serverAddr)
	if err != nil {
		return errors.Wrap(err, "cannot connect to server")
	}
	c.Conn = initialConnection

	fmt.Printf("Connection to server at %+v has been created\n", serverAddr)

	rrqPacket := packets.NewRRQPacket(requestedFilePath, packets.Netascii)
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

		// Parse the bytes received into a packet
		parsedPacket, err := packets.ParsePacket(buf)
		if err != nil {
			return errors.Wrap(err, "cannot parse incoming packet")
		}

		switch parsedPacket := parsedPacket.(type) {
		case packets.ErrorPacket:
			logger.Info("Error packet with following content has been received: %+v", parsedPacket)
			return nil
		case packets.DataPacket:
			receivedBytes = append(receivedBytes, parsedPacket.Data...)
			if len(parsedPacket.Data) < utils.MAX_DATA_FIELD_LENGTH {
				isFinalBlock = true
			}

			ackPacket := packets.NewAckPacket(parsedPacket.BlockNumber)

			_, err = newConnection.WriteToUDP(ackPacket.Bytes(), remoteAddr)
			if err != nil {
				return errors.Wrap(err, "cannot write to server")
			}
		}
	}

	splittedFilePath := strings.Split(requestedFilePath, "/")
	receivedFile := splittedFilePath[len(splittedFilePath)-1]
	f, err := os.Create(receivedFile)
	if err != nil {
		return errors.Wrap(err, "cannot create file to be received")
	}
	f.WriteString(string(receivedBytes))

	return nil
}

func (c *Client) WriteFile(serverAddr *net.UDPAddr, fileToWritePath string) error {
	var localAddress *net.UDPAddr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: c.TID}

	logger.Debug("The client local address is %+v", localAddress)

	initialConnection, err := net.DialUDP("udp4", localAddress, serverAddr)
	if err != nil {
		return errors.Wrap(err, "cannot connect to server")
	}
	c.Conn = initialConnection

	logger.Debug("Connection to server at %+v has been created", serverAddr)

	wrqPacket := packets.NewWRQPacket(fileToWritePath, packets.Netascii)
	_, err = c.Conn.Write(wrqPacket.Bytes())
	if err != nil {
		return errors.Wrap(err, "cannot write to server")
	}

	logger.Debug("Client has sent the first WRQ packet to the server")
	initialConnection.Close()

	// Listen for incoming connection from the server
	newConnection, err := net.ListenUDP("udp4", localAddress)
	if err != nil {
		return errors.Wrap(err, "error while listening for incoming UDP connections")
	}
	// defer newConnection.Close()

	logger.Debug("New connection has been created")

	logger.Info(">>> Reading requested file from the file-system: %s", wrqPacket.Filename)
	fileToWriteContent, err := utils.ReadFileFromFS(wrqPacket.Filename)
	if err != nil {
		errorPacket := packets.NewErrorPacket(2, err.Error())
		_, err = newConnection.Write(errorPacket.Bytes())
		if err != nil {
			fmt.Println(err)
			return errors.Wrapf(err, "cannot send error packet to server %+v", serverAddr)
		}
		return errors.Wrap(err, "cannot read requested file from server FS")
	}

	fileDataBlocks := utils.CreateDataBlocks(fileToWriteContent)

	for _, dataBlock := range fileDataBlocks {
		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		_, remoteAddress, err := newConnection.ReadFromUDP(buf)
		if err != nil {
			return errors.Wrap(err, "cannot read client request")
		}

		// Parse the bytes received into a packet
		parsedPacket, err := packets.ParsePacket(buf)
		if err != nil {
			return errors.Wrap(err, "cannot parse incoming packet")
		}

		switch parsedPacket := parsedPacket.(type) {
		case packets.AckPacket:
			dataPacket := packets.NewDataPacket(parsedPacket.BlockNumber+1, dataBlock)
			_, err := newConnection.WriteToUDP(dataPacket.Bytes(), remoteAddress)
			if err != nil {
				logger.Error("%+v", err)
				return errors.Wrapf(err, "cannot send data to machine %+v", serverAddr)
			}
		default:
			errorPacket := packets.NewErrorPacket(1, "invalid packet received")
			_, err = newConnection.Write(errorPacket.Bytes())
			if err != nil {
				logger.Error("%+v", err)
				return errors.Wrapf(err, "cannot send error packet to machine %+v", serverAddr)
			}
		}
	}
	return nil
}

package server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/mirkoschicchi/TFTP/internal/app/packets"
	"github.com/mirkoschicchi/TFTP/internal/app/utils"
	"github.com/pkg/errors"
)

type Server struct {
	TID    int
	Reader io.Reader
	Writer io.Writer
}

func (s *Server) Listen() error {
	localAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:69")
	if err != nil {
		return errors.Wrap(err, "cannot resolve the local UDP address")
	}

	initialConnection, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return errors.Wrap(err, "error while listening for incoming UDP connections")
	}
	defer initialConnection.Close()

	fmt.Printf("Server listening on %s\n", initialConnection.LocalAddr().String())

	for {
		fmt.Println("I am waiting to receive packets from clients")
		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		_, remoteAddr, err := initialConnection.ReadFromUDP(buf)
		if err != nil {
			return errors.Wrap(err, "cannot read client request")
		}

		parsedPacket, err := packets.ParsePacket(buf)
		if err != nil {
			return errors.Wrap(err, "cannot parse incoming packet")
		}
		go s.handleRRQRequest(remoteAddr, parsedPacket.(packets.RRQPacket))
	}

	return nil
}

func (s *Server) handleRRQRequest(clientAddr *net.UDPAddr, rrqPacket packets.RRQPacket) error {
	fmt.Printf(">>> Connecting to client having address %+v\n", clientAddr)

	randomTID := utils.GetRandomTID()
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", randomTID))
	if err != nil {
		return errors.Wrap(err, "cannot resolve the local UDP address")
	}

	fmt.Printf(">>> I have generated a random TID: %d\n", randomTID)

	newConnection, err := net.DialUDP("udp4", localAddr, clientAddr)
	if err != nil {
		return errors.Wrapf(err, "cannot instantiate new connection to machine %+v", clientAddr)
	}

	fmt.Println(">>> Reading requested file from the file-system")
	requestedFileContent, err := s.ReadFileFromFS(rrqPacket.Filename)
	if err != nil {
		return errors.Wrap(err, "cannot read requested file from server FS")
	}

	fileDataBlocks := createDataBlocks(requestedFileContent)

	for blockCounter, dataBlock := range fileDataBlocks {
		dataPacket := packets.NewDataPacket(uint16(blockCounter+1), dataBlock)
		_, err = newConnection.Write(dataPacket.Bytes())
		if err != nil {
			fmt.Println(err)
			return errors.Wrapf(err, "cannot send data to machine %+v", clientAddr)
		}

		// fmt.Printf(">>> The data packet has been sent to the client: %+v\n", dataPacket)

		// fmt.Println(">>> I am waiting to receive packets from clients")
		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		_, _, err := newConnection.ReadFromUDP(buf)
		if err != nil {
			return errors.Wrap(err, "cannot read client request")
		}

		_, err = packets.ParsePacket(buf)
		if err != nil {
			return errors.Wrap(err, "cannot parse incoming packet")
		}
		// fmt.Printf(">>> I have received the following packet: %+v\n", parsedPacket)
	}

	return nil
}

func (s *Server) ReadFileFromFS(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "cannot read file %s", filename)
	}

	return content, nil
}

// createDataBlocks returns a list of bytes array splitted in blocks
// of size 512
func createDataBlocks(fileContent []byte) [][]byte {
	numberOfBlocks := utils.CalculateNumberOfBlocks(len(fileContent))

	var dataBlocks [][]byte
	for i := 0; i < numberOfBlocks; i++ {
		if i == numberOfBlocks-1 {
			dataBlocks = append(dataBlocks, fileContent[512*i:])
			continue
		}
		dataBlocks = append(dataBlocks, fileContent[512*i:512*(i+1)])
	}

	return dataBlocks
}

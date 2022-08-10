package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mirkoschicchi/TFTP/internal/app/logger"
	"github.com/mirkoschicchi/TFTP/internal/app/packets"
	"github.com/mirkoschicchi/TFTP/internal/app/utils"
	"github.com/pkg/errors"
)

type Server struct {
	Wg *sync.WaitGroup
}

func NewServer() *Server {
	server := new(Server)
	server.Wg = new(sync.WaitGroup)

	return server
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

	logger.Info("Server listening on %s", initialConnection.LocalAddr().String())

	signalChannel := make(chan os.Signal)
	quitChannel := make(chan bool)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		logger.Warning("CTRL-C has been pressed. Shutting down the server")
		initialConnection.Close()
		quitChannel <- true
	}()

LOOP:
	for {
		select {
		case <-quitChannel:
			break LOOP
		default:
			logger.Info("Server is waiting to receive packets from clients")
			var buf []byte = make([]byte, packets.TftpMaxPacketSize)
			_, remoteAddr, err := initialConnection.ReadFromUDP(buf)
			if err != nil {
				return errors.Wrap(err, "cannot read client request")
			}

			parsedPacket, err := packets.ParsePacket(buf)
			if err != nil {
				return errors.Wrap(err, "cannot parse incoming packet")
			}

			switch parsedPacket := parsedPacket.(type) {
			case packets.RRQPacket:
				s.Wg.Add(1)
				go s.handleRRQRequest(remoteAddr, parsedPacket)
			case packets.WRQPacket:
				s.Wg.Add(1)
				go s.handleWRQRequest(remoteAddr, parsedPacket)
			default:
				logger.Warning("Unexpected packet received. Ignoring it")
			}
		}
	}

	s.Wg.Wait()
	logger.Info("The server and related go-routines has been shutted down")
	return nil
}

func (s *Server) handleRRQRequest(clientAddr *net.UDPAddr, rrqPacket packets.RRQPacket) error {
	defer s.Wg.Done()
	logger.Info(">>> Client having address %+v has requested to read file %s", clientAddr, rrqPacket.Filename)

	randomTID := utils.GetRandomTID()
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", randomTID))
	if err != nil {
		return errors.Wrap(err, "cannot resolve the local UDP address")
	}

	logger.Debug(">>> Server has generated a random TID: %d", randomTID)

	logger.Debug(">>> Server is creating a new connection to the client using local port %d", randomTID)
	newConnection, err := net.DialUDP("udp4", localAddr, clientAddr)
	if err != nil {
		return errors.Wrapf(err, "cannot instantiate new connection to machine %+v", clientAddr)
	}
	defer newConnection.Close()

	logger.Debug(">>> Reading requested file from the file-system: %s", rrqPacket.Filename)
	requestedFileContent, err := utils.ReadFileFromFS(rrqPacket.Filename)
	if err != nil {
		logger.Error("Cannot read file %s", rrqPacket.Filename)
		errorPacket := packets.NewErrorPacket(1, fmt.Sprintf("File %s has not been found in the server. Err: %v", rrqPacket.Filename, err))
		_, err = newConnection.Write(errorPacket.Bytes())
		if err != nil {
			logger.Error("%+v", err)
			return errors.Wrapf(err, "cannot send error packet to client %+v", clientAddr)
		}
		return errors.Wrap(err, "cannot read requested file from server FS")
	}

	// Split the file in blocks of max length 512 bytes
	fileDataBlocks, numberOfBlocks := utils.CreateDataBlocks(requestedFileContent)
	logger.Debug(">>> The file has been splitted into %d blocks", numberOfBlocks)

	for blockCounter, dataBlock := range fileDataBlocks {
		dataPacket := packets.NewDataPacket(uint16(blockCounter+1), dataBlock)
		bytesWritten, err := newConnection.Write(dataPacket.Bytes())
		if err != nil {
			logger.Error("%+v", err)
			return errors.Wrapf(err, "cannot send data to machine %+v", clientAddr)
		}
		logger.Debug(">>> The server has sent %d bytes to the client", bytesWritten)

		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		bytesReceived, _, err := newConnection.ReadFromUDP(buf)
		if err != nil {
			logger.Error("%+v", err)
			return errors.Wrap(err, "cannot read client request")
		}
		logger.Debug("The server has received %d bytes from the client", bytesReceived)

		_, err = packets.ParsePacket(buf)
		if err != nil {
			logger.Error("%+v", err)
			return errors.Wrap(err, "cannot parse incoming packet")
		}
	}
	return nil
}

func (s *Server) handleWRQRequest(clientAddr *net.UDPAddr, wrqPacket packets.WRQPacket) error {
	defer s.Wg.Done()
	logger.Info(">>> Client having address %+v has requested to write file %s", clientAddr, wrqPacket.Filename)

	randomTID := utils.GetRandomTID()
	localAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", randomTID))
	if err != nil {
		return errors.Wrap(err, "cannot resolve the local UDP address")
	}

	logger.Info(">>> Server has generated a random TID: %d", randomTID)

	newConnection, err := net.DialUDP("udp4", localAddr, clientAddr)
	newConnection.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		logger.Error("%+v", err)
		return errors.Wrapf(err, "cannot instantiate new connection to machine %+v", clientAddr)
	}
	logger.Debug("Server has initiated a new connection to the client using local port %d", randomTID)

	// Create initial ACK packet
	ackPacket := packets.NewAckPacket(0)
	_, err = newConnection.Write(ackPacket.Bytes())
	if err != nil {
		logger.Error("%+v", err)
		return errors.Wrapf(err, "cannot send initial ACK packet to client %+v", clientAddr)
	}

	var receivedBytes []byte
	var isFinalBlock bool = false

	for !isFinalBlock {
		var buf []byte = make([]byte, packets.TftpMaxPacketSize)
		_, _, err := newConnection.ReadFromUDP(buf)
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
			logger.Error("Error packet with following content has been received: %+v", parsedPacket)
			return nil
		case packets.DataPacket:
			receivedBytes = append(receivedBytes, parsedPacket.Data...)
			if len(parsedPacket.Data) < utils.MAX_DATA_FIELD_LENGTH {
				isFinalBlock = true
			}

			ackPacket := packets.NewAckPacket(parsedPacket.BlockNumber)

			_, err := newConnection.Write(ackPacket.Bytes())
			if err != nil {
				return errors.Wrap(err, "cannot write to server")
			}
		}

	}

	splittedFilePath := strings.Split(wrqPacket.Filename, "/")
	receivedFile := splittedFilePath[len(splittedFilePath)-1]
	logger.Debug("Saving received file to file-system. File location: %s", receivedFile)
	f, err := os.Create(receivedFile)
	if err != nil {
		return errors.Wrap(err, "cannot create file to be received")
	}
	f.WriteString(string(receivedBytes))

	return nil
}

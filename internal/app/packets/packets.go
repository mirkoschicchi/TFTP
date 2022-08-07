package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

const TftpMaxPacketSize = 1468

// Packet represents any TFTP packet
type Packet interface {
	// GetType returns the packet type
	GetType() uint16

	// Bytes serializes the packet
	Bytes() []byte
}

type Mode string

const (
	opRRQ   = uint16(1) // Read request (RRQ)
	opWRQ   = uint16(2) // Write request (WRQ)
	opDATA  = uint16(3) // Data
	opACK   = uint16(4) // Acknowledgement
	opERROR = uint16(5) // Error
)

const (
	Netascii Mode = "netascii"
	Octet    Mode = "octet"
	Mail     Mode = "mail"
)

type RRQPacket struct {
	Opcode   uint16
	Filename string
	Mode     Mode
}

type WRQPacket struct {
	Opcode   uint16
	Filename string
	Mode     Mode
}

type DataPacket struct {
	Opcode      uint16
	BlockNumber uint16
	Data        []byte
}

type AckPacket struct {
	Opcode      uint16
	BlockNumber uint16
}

type ErrorPacket struct {
	Opcode    uint16
	ErrorCode uint16
	ErrMsg    string
}

func NewRRQPacket(filename string, mode Mode) RRQPacket {
	return RRQPacket{Opcode: opRRQ, Filename: filename, Mode: mode}
}

func NewWRQPacket(filename string, mode Mode) WRQPacket {
	return WRQPacket{Opcode: opWRQ, Filename: filename, Mode: mode}
}

func NewDataPacket(blockNumber uint16, data []byte) DataPacket {
	return DataPacket{Opcode: opDATA, BlockNumber: blockNumber, Data: data}
}

func NewAckPacket(blockNumber uint16) AckPacket {
	return AckPacket{Opcode: opACK, BlockNumber: blockNumber}
}

func NewErrorPacket(errorCode uint16, errMsg string) ErrorPacket {
	return ErrorPacket{Opcode: opERROR, ErrorCode: errorCode, ErrMsg: errMsg}
}

func (rrqPacket RRQPacket) Bytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.BigEndian.PutUint16(encodedPacket, rrqPacket.Opcode)
	encodedPacket = append(encodedPacket, []byte(rrqPacket.Filename)...)
	encodedPacket = append(encodedPacket, byte(0))
	encodedPacket = append(encodedPacket, []byte(rrqPacket.Mode)...)
	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

func (wrqPacket WRQPacket) Bytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.BigEndian.PutUint16(encodedPacket, wrqPacket.Opcode)
	encodedPacket = append(encodedPacket, []byte(wrqPacket.Filename)...)
	encodedPacket = append(encodedPacket, byte(0))
	encodedPacket = append(encodedPacket, []byte(wrqPacket.Mode)...)
	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

func (dataPacket DataPacket) Bytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.BigEndian.PutUint16(encodedPacket, dataPacket.Opcode)

	encodedBlockNumber := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedBlockNumber, dataPacket.BlockNumber)
	encodedPacket = append(encodedPacket, encodedBlockNumber...)

	encodedPacket = append(encodedPacket, (dataPacket.Data)...)

	return encodedPacket
}

func (ackPacket AckPacket) Bytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.BigEndian.PutUint16(encodedPacket, ackPacket.Opcode)

	encodedBlockNumber := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedBlockNumber, ackPacket.BlockNumber)
	encodedPacket = append(encodedPacket, encodedBlockNumber...)

	return encodedPacket
}

func (errorPacket ErrorPacket) Bytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.BigEndian.PutUint16(encodedPacket, errorPacket.Opcode)

	encodedErrorCode := make([]byte, 2)
	binary.BigEndian.PutUint16(encodedErrorCode, errorPacket.ErrorCode)
	encodedPacket = append(encodedPacket, encodedErrorCode...)

	encodedPacket = append(encodedPacket, []byte(errorPacket.ErrMsg)...)

	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

func rrqPacketFromBytes(b []byte) RRQPacket {
	var parsedPacket RRQPacket

	vals := bytes.Split(b[2:], []byte{0})
	filename := vals[0]
	mode := vals[1]

	parsedPacket = NewRRQPacket(string(filename), Mode(mode))

	return parsedPacket
}

func wrqPacketFromBytes(b []byte) WRQPacket {
	var parsedPacket WRQPacket

	vals := bytes.Split(b[2:], []byte{0})
	filename := vals[0]
	mode := vals[1]

	parsedPacket = NewWRQPacket(string(filename), Mode(mode))

	return parsedPacket
}

func dataPacketFromBytes(b []byte) DataPacket {
	var parsedPacket DataPacket

	var blockNumber uint16
	blockNumber = binary.BigEndian.Uint16(b[2:4])
	data := b[4:]
	splitted := bytes.Split(data, []byte{0})
	parsedPacket = NewDataPacket(uint16(blockNumber), splitted[0])

	return parsedPacket
}

func ackPacketFromBytes(b []byte) AckPacket {
	var parsedPacket AckPacket

	var blockNumber uint16
	blockNumber = binary.BigEndian.Uint16(b[2:4])

	parsedPacket = NewAckPacket(uint16(blockNumber))

	return parsedPacket
}

func errorPacketFromBytes(b []byte) ErrorPacket {
	var parsedPacket ErrorPacket

	var errorCode uint16
	errorCode = binary.BigEndian.Uint16(b[2:4])
	errorMsg := b[4:]

	parsedPacket = NewErrorPacket(errorCode, string(errorMsg))

	return parsedPacket
}

func ParsePacket(p []byte) (interface{}, error) {
	packetLength := len(p)

	if packetLength < 2 {
		return nil, errors.New("Invalid packet: packet is too short")
	}

	opCode := binary.BigEndian.Uint16(p)
	switch opCode {
	case opRRQ:
		if packetLength < 4 {
			return nil, fmt.Errorf("short RRQ packet: %d", packetLength)
		}
		return rrqPacketFromBytes(p), nil
	case opWRQ:
		if packetLength < 4 {
			return nil, fmt.Errorf("short WRQ packet: %d", packetLength)
		}
		return wrqPacketFromBytes(p), nil
	case opDATA:
		if packetLength < 4 {
			return nil, fmt.Errorf("short DATA packet: %d", packetLength)
		}
		return dataPacketFromBytes(p), nil
	case opACK:
		if packetLength < 4 {
			return nil, fmt.Errorf("short ACK packet: %d", packetLength)
		}
		return ackPacketFromBytes(p), nil
	case opERROR:
		if packetLength < 5 {
			return nil, fmt.Errorf("short ERROR packet: %d", packetLength)
		}
		return errorPacketFromBytes(p), nil
	default:
		return nil, fmt.Errorf("unknown opcode: %d", opCode)

	}
}

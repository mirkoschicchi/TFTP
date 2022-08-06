package packets

import "encoding/binary"

type Mode string

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
	return RRQPacket{Opcode: 0x01, Filename: filename, Mode: mode}
}

func NewWRQPacket(filename string, mode Mode) WRQPacket {
	return WRQPacket{Opcode: 0x02, Filename: filename, Mode: mode}
}

func NewDataPacket(blockNumber uint16, data []byte) DataPacket {
	return DataPacket{Opcode: 0x03, BlockNumber: blockNumber, Data: data}
}

func NewAckPacket(blockNumber uint16) AckPacket {
	return AckPacket{Opcode: 0x04, BlockNumber: blockNumber}
}

func NewErrorPacket(errorCode uint16, errMsg string) ErrorPacket {
	return ErrorPacket{Opcode: 0x05, ErrorCode: errorCode, ErrMsg: errMsg}
}

func (rrqPacket RRQPacket) ToBytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.LittleEndian.PutUint16(encodedPacket, rrqPacket.Opcode)
	encodedPacket = append(encodedPacket, []byte(rrqPacket.Filename)...)
	encodedPacket = append(encodedPacket, byte(0))
	encodedPacket = append(encodedPacket, []byte(rrqPacket.Mode)...)
	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

func (wrqPacket WRQPacket) ToBytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.LittleEndian.PutUint16(encodedPacket, wrqPacket.Opcode)
	encodedPacket = append(encodedPacket, []byte(wrqPacket.Filename)...)
	encodedPacket = append(encodedPacket, byte(0))
	encodedPacket = append(encodedPacket, []byte(wrqPacket.Mode)...)
	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

func (dataPacket DataPacket) ToBytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.LittleEndian.PutUint16(encodedPacket, dataPacket.Opcode)

	encodedBlockNumber := make([]byte, 2)
	binary.LittleEndian.PutUint16(encodedBlockNumber, dataPacket.BlockNumber)
	encodedPacket = append(encodedPacket, encodedBlockNumber...)

	encodedPacket = append(encodedPacket, (dataPacket.Data)...)

	return encodedPacket
}

func (ackPacket AckPacket) ToBytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.LittleEndian.PutUint16(encodedPacket, ackPacket.Opcode)

	encodedBlockNumber := make([]byte, 2)
	binary.LittleEndian.PutUint16(encodedBlockNumber, ackPacket.BlockNumber)
	encodedPacket = append(encodedPacket, encodedBlockNumber...)

	return encodedPacket
}

func (errorPacket ErrorPacket) ToBytes() []byte {
	encodedPacket := make([]byte, 2)

	binary.LittleEndian.PutUint16(encodedPacket, errorPacket.Opcode)

	encodedErrorCode := make([]byte, 2)
	binary.LittleEndian.PutUint16(encodedErrorCode, errorPacket.ErrorCode)
	encodedPacket = append(encodedPacket, encodedErrorCode...)

	encodedPacket = append(encodedPacket, []byte(errorPacket.ErrMsg)...)

	encodedPacket = append(encodedPacket, byte(0))

	return encodedPacket
}

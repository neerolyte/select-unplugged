package sp

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type Request struct {
	message Message
}

func NewRequestFromMessage(message Message) Request {
	return Request{
		message: message,
	}
}

func NewRequestQuery(area Area) Request {
	message := []byte("Q")
	message = append(message, area.Message()...)
	message = binary.LittleEndian.AppendUint16(message, Crc(message))

	return Request{
		message: message,
	}
}

func NewRequestWrite(memory Memory) Request {
	message := Message("W")
	message = append(message, uint8(memory.Area().Words()-1))
	message = append(message, memory.Area().Address().Message()...)
	message = binary.LittleEndian.AppendUint16(message, Crc(message))
	message = append(message, memory.data...)
	message = binary.LittleEndian.AppendUint16(message, Crc(message))
	return Request{
		message: message,
	}
}

func (r Request) Message() Message {
	return r.message
}

func (r Request) String() string {
	return fmt.Sprintf("Request(0x%s)", hex.EncodeToString(r.Message()))
}

// Calculate request length given enough bytes from a Message
func CalculateRequestLength(partial Message) (*int, error) {
	i := func(i int) *int { return &i }
	if len(partial) < 1 {
		return nil, errors.New("Need a byte to calculate length")
	}
	rt := MessageType(partial[0])
	length := 8
	if rt == Query {
		return i(8), nil
	}
	if rt != Write {
		return nil, errors.New("Unknown message type")
	}
	words, err := partial.Words()
	if err != nil {
		return nil, err
	}
	return i(length + *words*2 + 2), nil
}

func (r Request) ResponseLength() (*int, error) {
	requestType, err := r.Type()
	if err != nil {
		return nil, err
	}
	requestLength := len(r.Message())
	dataLength := r.DataLength()
	if err != nil {
		return nil, err
	}
	crcLength := 2
	if *requestType == Write {
		return &requestLength, nil
	}
	length := requestLength + dataLength + crcLength
	return &length, nil
}

func (r Request) DataLength() int {
	return (int(r.message[1]) + 1) * 2
}

func (r Request) Type() (*MessageType, error) {
	return r.Message().Type()
}

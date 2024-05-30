package messages

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MsgChoke         = 0
	MsgUnchoke       = 1
	MsgInterested    = 2
	MsgNotInterested = 3
	MsgHave          = 4
	MsgBitfield      = 5
	MsgRequest       = 6
	MsgPiece         = 7
	MsgCancel        = 8
)

type Message struct {
	Id      byte
	Payload []byte
}

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{Id: MsgRequest, Payload: payload}
}

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.Id != MsgPiece {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MsgPiece, msg.Id)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(msg.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(buf))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.Id)
	copy(buf[5:], m.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		Id:      messageBuf[0],
		Payload: messageBuf[1:],
	}

	return &m, nil
}

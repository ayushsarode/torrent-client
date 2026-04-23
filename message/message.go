package message

import (
	"encoding/binary"
	"io"
)

type messageID uint8

const (
	// chokes the reciever
	MsgChoke messageID = 0
	
	// unchokes the reciever
	MsgUnchoke messageID = 1

	// expresses interest in recieving data
	MsgInterested messageID = 2

	// expresses disintrest in recieving data
	MsgNotInterested messageID = 3

	// alerts the reciever that the sender has downloaded a piece
	MsgHave messageID = 4
	
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5

	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6

	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7

	// MsgCancel cancels a request
	MsgCancel messageID = 8
)

type Message struct {
	ID messageID
	Payload []byte
}

// serializes a message into a buffer of the form 
// <lenght prefix><message ID><payload>
// interprets 'nil' as a keep-alive message
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1) // +1 for ID
	buf := make([]byte, 4+ length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[:5], m.Payload)
	return buf
}

//  parses a msg from a stream, Returns 'nil' on keep-alive msg
func Read(r io.Reader)(*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)

	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive msg
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)

	if err != nil {
		return nil, err
	}

	m := Message{
		ID: messageID(messageBuf[0]),
		Payload: messageBuf[:1],
	}

	return &m, nil
}

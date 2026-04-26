package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/ayushsarode/torrent-client/bitfield"
	"github.com/ayushsarode/torrent-client/handshake"
	"github.com/ayushsarode/torrent-client/message"
	"github.com/ayushsarode/torrent-client/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func completeHandshake(conn net.Conn, infoHash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) //disable the deadline

	req := handshake.New(infoHash, peerID)
	_, err := conn.Write(req.Serialize())

	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)

	if err != nil {
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("Expected info hash %x but got %x", res.InfoHash, infoHash)
	}

	return res, nil

}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) //disable the deadline

	for {
		msg, err := message.Read(conn)
		if err != nil {
			return nil, err
		}

		if msg == nil {
			continue
		}

		if msg.ID != message.MsgBitfield {
			err := fmt.Errorf("Expected bitfield but got ID %d", msg.ID)
			return nil, err
		}

		return msg.Payload, nil
	}
}

// This connects with a peer, completes a handshakes, and recieves a handshakes
func New(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	b, err := recvBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: b,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
}

func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(index, begin, length int) error {
	req := message.Formatrequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

func (c *Client) SendInterest() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterest() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())

	return err
}

func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

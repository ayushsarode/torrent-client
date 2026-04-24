package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	Ptsr	string
	InfoHash	[20]byte
	PeerID	[20]byte
}

func New(InfoHash, PeerID [20]byte) *Handshake {
	return &Handshake{
		Ptsr: "BitTorrent protocol",
		InfoHash: InfoHash,
		PeerID: PeerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, len(h.Ptsr)+49)
	buf[0] = byte(len(h.Ptsr))
	curr := 1
	curr += copy(buf[curr:], []byte(h.Ptsr))
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], []byte(h.InfoHash[:]))
	curr += copy(buf[curr:], []byte(h.PeerID[:]))
	return buf
}

func Read(r io.Reader) (*Handshake, error) {
	lengthbuf:= make([]byte, 1)

	_, err := io.ReadFull(r, lengthbuf)

	if err != nil {
		return nil ,err
	}

	pstrlen := int(lengthbuf[0])
	

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen can't be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, pstrlen+48)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	h:= Handshake{
		Ptsr: string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID: peerID,
	}

	return &h, nil


}

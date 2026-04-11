package peers

import (
	"encoding/binary"
	"errors"
	"net"
)

type Peer struct {
	IP	net.IP
	Port	uint16
}

// this func parses peer IP addr and ports from a buffer.
// basically convert binary format to struct, that means we cant directly dial to http with these bytes.
func Unmarshal(peersBin []byte)([]Peer, error) {
	const peerSize = 6 // 4 byte for IP and 2 byte for port

	numPeers := len(peersBin) / peerSize

	if len(peersBin) % peerSize != 0 {
		err := errors.New("Received malformed peers")
		return nil, err
	}

	peers := make([]Peer, numPeers)

	for i :4= 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset + 4])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offse+6])
	}
	return peers, nil
}

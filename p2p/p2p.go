package p2p

import (
	"github.com/ayushsarode/torrent-client/client"	
	"github.com/ayushsarode/torrent-client/message"
	"github.com/ayushsarode/torrent-client/peers"
)

// it is the largefst number of bytes a req can ask for 
const MaxBlockSize = 16384

//it is the num. of unfulfilled reqs a client can have in its pipeline
const MaxBacklog = 5

type Torrent struct {
	Peers	[]peers.Peer
	PeerID	[20]byte
	InfoHash [20]byte
	PieceHashes int
	Length int
}




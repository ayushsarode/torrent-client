package torrentfile

import (
	"net/url"
	"strconv"
)

func (t *Torrentfile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Annouce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash": []string{string(t.InfoHash[:])},
		"peer_id": []string{string(peerID[:])},
		"port": []string{string(port)},
		"uploaded": []string{"0"},
		"downloaded": []string{"0"},
		"compact": []string{"1"},
		"left": []string{strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *Torrentfile) requestPeers(peerID [20] byte, port uint16) ([]Peer, error) {
	const peerSize = 6;
	numPeers := len()
}

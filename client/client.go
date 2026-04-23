package client

import (
	"net"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second) 

	if err != nil {
		return nil, err
	}
}

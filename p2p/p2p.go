package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/ayushsarode/torrent-client/client"
	"github.com/ayushsarode/torrent-client/message"
	"github.com/ayushsarode/torrent-client/peers"
)

// it is the largefst number of bytes a req can ask for
const MaxBlockSize = 16384

// it is the num. of unfulfilled reqs a client can have in its pipeline
const MaxBacklog = 5

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()

	if err != nil {
		return err
	}

	if msg == nil { //keep alive
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}

		state.downloaded += n
		state.backlog--

	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index: pw.index,
		client: c,
		buf: make([]byte, pw.length),
	}

	// deadline helps get unresponsive peers unstuck.
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))

	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		// if unchoked, send requests until we have enough unfulfilled requests
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize

				// last block might be shorter than the typical block
				// this prevents over-requesting and ensures the last block fits the piece exactly
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					 return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}

// it compares the SHA-1 of the downloaded piece against the expected piece hash from torrent metadata
func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)

	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed to integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult, workerDone chan<- struct{}) {
	defer func() {
		workerDone <- struct{}{}
	}()

	c, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		return
	}

	defer c.Conn.Close()

	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterest()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.index) {
			workQueue <- pw // put the piece back so another worker/peer can try it
			continue
		}

		buf, err := attemptDownloadPiece(c, pw)

		if err != nil {
			log.Println("exiting", err)
			workQueue <- pw // put piece back on the queue
			return
		}

		err = checkIntegrity(pw,buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQueue <- pw //put piece back on the queue
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength

	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)
	if len(t.Peers) == 0 {
		return nil, fmt.Errorf("no peers available for download")
	}

	// init queues for workers to retrive work and send results
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	workerDone := make(chan struct{}, len(t.Peers))

	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// start workers

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results, workerDone)
	}

	//collect results into a buffer until full
	buf := make([]byte, t.Length)

	donePieces := 0
	activeWorkers := len(t.Peers)
	stallTimer := time.NewTimer(45 * time.Second)
	defer stallTimer.Stop()

	for donePieces < len(t.PieceHashes) {
		if activeWorkers == 0 {
			return nil, fmt.Errorf("all peers disconnected before download completed (%d/%d pieces)", donePieces, len(t.PieceHashes))
		}

		select {
		case res := <-results:
			begin, end := t.calculateBoundsForPiece(res.index)
			copy(buf[begin:end], res.buf)
			donePieces++

			percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
			numWorkers := runtime.NumGoroutine() - 1 // sub 1 for main thread
			log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)

			if !stallTimer.Stop() {
				select {
				case <-stallTimer.C:
				default:
				}
			}
			stallTimer.Reset(45 * time.Second)
		case <-workerDone:
			activeWorkers--
		case <-stallTimer.C:
			return nil, fmt.Errorf("download stalled: no piece progress for 45s (%d/%d pieces)", donePieces, len(t.PieceHashes))
		}
	}
	close(workQueue)

	return buf, nil
}

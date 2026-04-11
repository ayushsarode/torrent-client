package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/jackpal/bencode-go"
)


type Torrentfile struct {
	Annouce string
	InfoHash [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length int
	Name string
}

type bencodeInfo struct {
	Pieces string `bencode:"pieces"`
	PieceLength int `bencode:"piece length"`
	Length int `bencode:"length"`
	Name string `bencode:"name"`
}

type bencodeTorrent struct {
	Annouce string `bencode:"announce"`
	Info bencodeInfo `bencode:"info"`
}

func Open(path string) (Torrentfile, error)  {
	file, err := os.Open(path)

	if err != nil {
		return Torrentfile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}

	err = bencode.Unmarshal(file,err)

	if  err != nil {
		return Torrentfile{}, err
	}
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer

	err := bencode.Marshal(&buf, *i)

	if err != nil {
		return [20]byte{}, err
	}

	h := sha1.Sum(buf.Bytes())

	return h, nil
} 

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20

	buf := []byte(i.Pieces)

	if len(buf) % hashLen != 0 {
		err := fmt.Errorf("Received malformed pieces of lenght %d", len(buf))
		return nil
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes) 

	fori i := 0;i < numHashes; i++ {
		copy
	}

}

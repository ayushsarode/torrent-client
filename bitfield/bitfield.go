package bitfield

import ()

type Bitfield []byte

// HasPiece tells if a bitfield has a particular index sets
func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	return b[byteIndex]>>(7-offset)&1 != 0
}

// Setpiece sets a bit in the bitfield
func(b Bitfield) SetPiece(index int) {
	byteIndex := index/8
	offset := index % 8
	b[byteIndex] |=  1 << (7 - offset)
}





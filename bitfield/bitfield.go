package bitfield

type Bitfield []byte

// HasPiece tells if a bitfield has a particular index sets
func (b Bitfield) HasPiece(index int) bool {
	if index < 0 {
		return false
	}
	byteIndex := index / 8
	if byteIndex >= len(b) {
		return false
	}
	offset := index % 8
	return b[byteIndex]>>(7-offset)&1 != 0
}

// Setpiece sets a bit in the bitfield
func(b Bitfield) SetPiece(index int) {
	if index < 0 {
		return
	}
	byteIndex := index/8
	if byteIndex >= len(b) {
		return
	}
	offset := index % 8
	b[byteIndex] |=  1 << (7 - offset)
}





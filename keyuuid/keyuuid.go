package keyuuid

import (
	"encoding/binary"

	"github.com/google/uuid"
)

// Value is the encoded form—a standard 16-byte UUID.
type Value = uuid.UUID

// Encoder mirrors key32/key64: it carries the bit-layout and does Encode/Decode.
type Encoder interface {
	Encode(v uuid.UUID) Value
	Decode(v Value) uuid.UUID
	Prefix(v Value) uuid.UUID // top bits, still as a UUID (with other bits zeroed)
	LeftSize() int            // bits to the right of the prefix
	PrefixSize() int          // number of bits in the prefix
	RightSize() int           // bits left of the prefix
}

// encoder is the concrete
type encoder struct {
	totalBits  int // 0 means “identity over 128 bits”, otherwise ≤64
	maskOffset int // offset within that field
	prefixSize int // how many bits to extract & reverse
}

// NewEncoder lets you make any UUID‐based encoder.
// totalBits ≤ 64, maskOffset+prefixSize ≤ totalBits.
func NewEncoder(totalBits, maskOffset, prefixSize int) Encoder {
	return encoder{totalBits, maskOffset, prefixSize}
}

// NewUUIDv7Encoder: extract the top 48 bits as timestamp, reverse 4 bits at offset 11
func NewUUIDv7Encoder() Encoder {
	return encoder{48, 11, 4}
}

// NewULIDEncoder: ULID also puts its 48-bit timestamp in the top 48 bits,
// and we reverse the low 16 of that if you like (or pick any shard size).
func NewULIDEncoder() Encoder {
	return encoder{48 /*shard offset*/, 16 /*shard size*/, 16}
}

func (e encoder) LeftSize() int {
	if e.totalBits == 0 {
		return 128
	}
	return e.totalBits - e.maskOffset - e.prefixSize
}

func (e encoder) PrefixSize() int { return e.prefixSize }

func (e encoder) RightSize() int {
	if e.totalBits == 0 {
		return 0
	}
	return e.maskOffset
}

// Encode plucks off the prefixSize bits starting at maskOffset within the top totalBits,
// reverses them, and prepends into a standard UUID’s MSB.
func (e encoder) Encode(u uuid.UUID) Value {
	// identity over full 128 bits?
	if e.totalBits == 0 && e.maskOffset == 0 && e.prefixSize == 0 {
		return u
	}
	// pull MSB=first 8 bytes as uint64
	msb := binary.BigEndian.Uint64(u[0:8])
	// isolate the “target” field (the top totalBits of msb)
	target := msb >> (64 - e.totalBits)

	mask := uint64((1 << e.prefixSize) - 1)
	field := (target >> e.maskOffset) & mask
	rev := reverseBits(field, e.prefixSize)

	left := (target >> (e.maskOffset + e.prefixSize)) // the high bits above the field
	right := target & ((1 << e.maskOffset) - 1)       // the low bits below it

	// reassemble into a shifted-down 48-bit, then back into msb
	encoded48 := (rev << (e.totalBits - e.prefixSize)) |
		(left << (e.maskOffset)) |
		right
	newMSB := (encoded48 << (64 - e.totalBits)) |
		(msb & ((1 << (64 - e.totalBits)) - 1))

	var out uuid.UUID
	binary.BigEndian.PutUint64(out[0:8], newMSB)
	copy(out[8:], u[8:])
	return out
}

// Decode inverts Encode.
func (e encoder) Decode(u uuid.UUID) uuid.UUID {
	// identity over full 128 bits?
	if e.totalBits == 0 && e.maskOffset == 0 && e.prefixSize == 0 {
		return u
	}

	msb := binary.BigEndian.Uint64(u[0:8])
	target := msb >> (64 - e.totalBits)

	mask := uint64((1 << e.prefixSize) - 1)
	rev := (target >> (e.totalBits - e.prefixSize)) & mask
	field := reverseBits(rev, e.prefixSize)

	left := (target >> e.maskOffset) & ((1 << (e.totalBits - e.maskOffset - e.prefixSize)) - 1)
	right := target & ((1 << e.maskOffset) - 1)

	decoded48 := (left << (e.maskOffset + e.prefixSize)) |
		(field << e.maskOffset) |
		right
	newMSB := (decoded48 << (64 - e.totalBits)) |
		(msb & ((1 << (64 - e.totalBits)) - 1))

	var out uuid.UUID
	binary.BigEndian.PutUint64(out[0:8], newMSB)
	copy(out[8:], u[8:])
	return out
}

// Prefix returns the high prefixSize bits of the encoded UUID (others zeroed).
func (e encoder) Prefix(u uuid.UUID) uuid.UUID {
	// identity => zero prefix
	if e.totalBits == 0 && e.maskOffset == 0 && e.prefixSize == 0 {
		return uuid.UUID{}
	}

	msb := binary.BigEndian.Uint64(u[0:8])
	mask := uint64((1 << e.prefixSize) - 1)
	top := (msb >> (64 - e.prefixSize)) & mask
	newMSB := top << (64 - e.prefixSize)

	var out uuid.UUID
	binary.BigEndian.PutUint64(out[0:8], newMSB)
	// rest stays zero
	return out
}

func reverseBits(x uint64, bitCount int) uint64 {
	var out uint64
	for i := 0; i < bitCount; i++ {
		out <<= 1
		out |= (x >> i) & 1
	}
	return out
}

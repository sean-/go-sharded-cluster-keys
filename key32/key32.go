// Package key32 provides a reusable Encoder for shuffling bits
// of a 32-bit value: extracting a run of bits, reversing them,
// and prepending them into the high bits of a 32-bit word.
package key32

// Value is the encoded form produced by Encoder.Encode.
type Value uint32

// Encoder defines the encode/decode interface and bit-layout metadata.
type Encoder interface {
	// Encode embeds v by extracting [offset..offset+size) bits,
	// reversing them, and prepending into the top size bits.
	Encode(v uint32) Value

	// Decode is the inverse of Encode, returning the original uint32.
	Decode(e Value) uint32

	// Prefix extracts the top size bits of e (the reversed segment).
	Prefix(e Value) uint32

	// PrefixHexPad takes a uint32 prefix and pads the LSB to hex-aligned nibbles.
	PrefixHexPad(p uint32) uint32

	// PrefixHexSize returns the number of hex nibbles required to display the prefix.
	PrefixHexSize() int

	// LeftSize is the number of LSB bits right of the prefix.
	LeftSize() int

	// PrefixSize is the width in bits of the prefix.
	PrefixSize() int

	// RightSize is the number of MSB bits left of the prefix.
	RightSize() int

	// EncodedBits returns the number of bits in Value
	EncodedBits() int
}

const valueBits = 32

// encoder is the concrete implementation of Encoder.
type encoder struct {
	offset    int // number of low bits to leave untouched
	size      int // size of the “shard” segment
	hexDigits int
}

// NewEncoder constructs an Encoder that will extract `size` bits
// starting at bit `offset` (0 = LSB), reverse them, and prepend into
// the top `size` bits.
func NewEncoder(offset, size int) Encoder {
	return encoder{
		offset:    offset,
		size:      size,
		hexDigits: (size + 3) / 4,
	}
}

func (e encoder) LeftSize() int   { return e.offset }
func (e encoder) PrefixSize() int { return e.size }
func (e encoder) RightSize() int  { return valueBits - e.offset - e.size }

// Encode implements Encoder.Encode
func (e encoder) Encode(v uint32) Value {
	// 1) extract the size-bit field
	mask := uint32((1 << e.size) - 1)
	field := (v >> e.offset) & mask

	// 2) reverse its bits
	rev := reverseBits(field, e.size)

	// 3) split out the untouched chunks
	left := v >> (e.offset + e.size)
	right := v & ((1 << e.offset) - 1)

	// 4) reassemble: [rev-pfx | left | right]
	enc := (rev << (valueBits - e.size)) | (left << e.offset) | right
	return Value(enc)
}

// Decode implements Encoder.Decode
func (e encoder) Decode(val Value) uint32 {
	u := uint32(val)
	mask := uint32((1 << e.size) - 1)

	// 1) pull out and unreversed the top size bits
	rev := (u >> (valueBits - e.size)) & mask
	field := reverseBits(rev, e.size)

	// 2) split the rest
	left := (u >> e.offset) & ((1 << (valueBits - e.size - e.offset)) - 1)
	right := u & ((1 << e.offset) - 1)

	// 3) rebuild original
	return (left << (e.offset + e.size)) | (field << e.offset) | right
}

// Prefix implements Encoder.Prefix
func (e encoder) Prefix(val Value) uint32 {
	u := uint32(val)
	return (u >> (valueBits - e.size)) & ((1 << e.size) - 1)
}

// PrefixHexPad implements Encoder.PrefixHexPad
func (e encoder) PrefixHexPad(prefix uint32) uint32 {
	// shift prefix so its MSB lands at the MSB of the nibble block
	return prefix << (e.hexDigits*4 - e.size)
}

// EncodedBits returns the number of bits in Value
func (e encoder) EncodedBits() int {
	return valueBits
}

// PrefixHexSize returns the number of hex nibbles required for a given prefix
func (e encoder) PrefixHexSize() int {
	return (e.size + 3) / 4
}

// reverseBits reverses the low bitCount bits of x.
func reverseBits(x uint32, bitCount int) uint32 {
	var out uint32
	for i := 0; i < bitCount; i++ {
		out <<= 1
		out |= (x >> i) & 1
	}
	return out
}

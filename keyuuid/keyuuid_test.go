package keyuuid

import (
	"encoding/binary"
	"testing"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestUUIDv7Encoder(t *testing.T) {
	// example UUIDv7 from oklog/ulid
	u7 := uuid.MustParse("018f14e0-8f0a-7def-91b4-f0ecb69f5f01")

	enc := NewUUIDv7Encoder()

	// metadata
	require.Equal(t, 11, enc.RightSize(), "RightSize")
	require.Equal(t, 4, enc.PrefixSize(), "PrefixSize")
	require.Equal(t, 48-11-4, enc.LeftSize(), "LeftSize")
	require.Equal(t, 48, enc.LeftSize()+enc.PrefixSize()+enc.RightSize(), "sum to totalBits")

	// round-trip
	encU := enc.Encode(u7)
	decU := enc.Decode(encU)
	require.Equal(t, u7, decU)

	// prefix: extract the 4-bit shard, reversed and prepended into top 4 bits
	// compute expected prefix manually:
	msb := binary.BigEndian.Uint64(u7[0:8])
	// timestamp48:
	ts48 := msb >> 16
	// field @ offset=11, size=4:
	field := (ts48 >> 11) & 0xF
	// reversed bits:
	var rev uint64
	for i := 0; i < 4; i++ {
		rev = (rev << 1) | ((field >> i) & 1)
	}
	// prefix in MSB: rev << (64-4)
	expectedTop := rev << 60
	// build a UUID with only those top bits
	var expPref uuid.UUID
	binary.BigEndian.PutUint64(expPref[0:8], expectedTop)
	gotPref := enc.Prefix(encU)
	require.Equal(t, expPref, gotPref)
}

func TestULIDEncoder(t *testing.T) {
	// example ULID
	uULID, err := ulid.Parse("01ARYZ6S41TSV4RRFFQ69G5FAV")
	require.NoError(t, err)

	// reinterpret as UUID bytes
	var base uuid.UUID
	copy(base[:], uULID[:])

	enc := NewULIDEncoder()

	// metadata: totalBits=48, offset=16, prefixSize=16
	require.Equal(t, 16, enc.RightSize(), "RightSize")
	require.Equal(t, 16, enc.PrefixSize(), "PrefixSize")
	require.Equal(t, 48-16-16, enc.LeftSize(), "LeftSize")
	require.Equal(t, 48, enc.LeftSize()+enc.PrefixSize()+enc.RightSize(), "sum to totalBits")

	// round-trip
	encU := enc.Encode(base)
	decU := enc.Decode(encU)
	require.Equal(t, base, decU)

	// prefix: extract 16-bit shard from ts48@offset16, reversed and prepended
	msb := binary.BigEndian.Uint64(base[0:8])
	ts48 := msb >> 16
	field := (ts48 >> 16) & 0xFFFF
	// reverse 16 bits
	var rev uint64
	for i := 0; i < 16; i++ {
		rev = (rev << 1) | ((field >> i) & 1)
	}
	expectedTop := rev << 48
	var expPref uuid.UUID
	binary.BigEndian.PutUint64(expPref[0:8], expectedTop)
	gotPref := enc.Prefix(encU)
	require.Equal(t, expPref, gotPref)
}

func TestGenericWrap(t *testing.T) {
	u := uuid.MustParse("3d813cbb-47fb-32ba-91df-831e1593ac29")
	enc := NewEncoder(0, 0, 0) // identity over full 128 bits
	// identity encode/decode
	encU := enc.Encode(u)
	require.Equal(t, u, encU)
	require.Equal(t, u, enc.Decode(encU))
	// prefix is zero UUID
	zero := uuid.UUID{}
	require.Equal(t, zero, enc.Prefix(encU))
	// metadata
	require.Equal(t, 128, enc.LeftSize(), "LeftSize for identity should be 128")
	require.Equal(t, 0, enc.PrefixSize(), "PrefixSize for identity should be 0")
	require.Equal(t, 0, enc.RightSize(), "RightSize for identity should be 0")
	require.Equal(t, 128, enc.LeftSize()+enc.PrefixSize()+enc.RightSize(),
		"sum of sizes should be 128 for identity",
	)
}

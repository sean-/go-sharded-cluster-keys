package key64

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder64_TableDriven(t *testing.T) {
	tests := []struct {
		name                         string
		orig                         uint64
		offset, size                 int
		wantLeft, wantPre, wantRight int
		wantEnc                      uint64
	}{
		{
			name:      "simple-8bit",
			orig:      0x0123456789ABCDEF,
			offset:    8,
			size:      8,
			wantLeft:  8,
			wantPre:   8,
			wantRight: 64 - 8 - 8,
			// field = (orig>>8)&0xFF = 0xCD
			// rev(0xCD,8) = 0xB3
			// enc = 0xB3<<56 | (orig>>16)<<8 | (orig&0xFF)
			wantEnc: 0xB30123456789ABEF,
		},
		{
			name:      "zero",
			orig:      0x0000000000000000,
			offset:    5,
			size:      4,
			wantLeft:  5,
			wantPre:   4,
			wantRight: 64 - 5 - 4,
			wantEnc:   0x0000000000000000,
		},
		{
			name:      "fullMask-32bit",
			orig:      0xFFFFFFFFFFFFFFFF,
			offset:    32,
			size:      32,
			wantLeft:  32,
			wantPre:   32,
			wantRight: 0,
			// field=0xFFFFFFFF, rev=0xFFFFFFFF, enc = all ones
			wantEnc: 0xFFFFFFFFFFFFFFFF,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			enc := NewEncoder(tc.offset, tc.size)

			// check size metadata
			require.Equal(t, tc.wantLeft, enc.LeftSize(), "LeftSize")
			require.Equal(t, tc.wantPre, enc.PrefixSize(), "PrefixSize")
			require.Equal(t, tc.wantRight, enc.RightSize(), "RightSize")
			require.Equal(t, 64,
				enc.LeftSize()+enc.PrefixSize()+enc.RightSize(),
				"sum of sizes should be 64",
			)

			// encoding
			gotEnc := enc.Encode(tc.orig)
			require.Equalf(
				t, tc.wantEnc, uint64(gotEnc),
				"Encode(0x%016x) = 0x%016x; want 0x%016x",
				tc.orig, gotEnc, tc.wantEnc,
			)

			// prefix extraction
			gotPre := enc.Prefix(gotEnc)
			mask := uint64((1 << tc.size) - 1)
			expPre := (uint64(gotEnc) >> (64 - tc.size)) & mask
			require.Equalf(t, expPre, gotPre,
				"Prefix(0x%016x) = 0x%x; want 0x%x",
				gotEnc, gotPre, expPre,
			)

			// decoding
			dec := enc.Decode(gotEnc)
			require.Equalf(
				t, tc.orig, dec,
				"Decode(0x%016x) = 0x%016x; want 0x%016x",
				gotEnc, dec, tc.orig,
			)
		})
	}
}

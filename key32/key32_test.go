package key32

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoderInterface_TableDriven(t *testing.T) {
	tests := []struct {
		name                         string
		orig                         uint32
		offset, size                 int
		wantLeft, wantPre, wantRight int
		wantEnc                      uint32
	}{
		{
			name:      "simple-8bit",
			orig:      0x12345678,
			offset:    8,
			size:      8,
			wantLeft:  8, // offset
			wantPre:   8, // size
			wantRight: 32 - 8 - 8,
			// field = (0x12345678>>8)&0xff = 0x56
			// rev(0x56,8)=0x6A
			// enc = 0x6A123478
			wantEnc: 0x6A123478,
		},
		{
			name:      "zero",
			orig:      0x00000000,
			offset:    5,
			size:      4,
			wantLeft:  5,
			wantPre:   4,
			wantRight: 32 - 5 - 4,
			wantEnc:   0x00000000,
		},
		{
			name:      "fullMask-16bit",
			orig:      0xFFFFFFFF,
			offset:    16,
			size:      16,
			wantLeft:  16,
			wantPre:   16,
			wantRight: 0,
			// field=0xffff, rev(0xffff,16)=0xffff, enc=0xffffffff
			wantEnc: 0xffffffff,
		},
		{
			name:      "65535-13bit",
			orig:      0x0000FFFF,
			offset:    11,
			size:      13,
			wantLeft:  11,
			wantPre:   13,
			wantRight: 32 - 11 - 13,
			// field=(0xffff>>11)&0x1fff = 0x1f
			// rev(0x1f,13)=0x1f00, enc=0x1f00<<19 | 0x07ff == 0xf80007ff
			wantEnc: 0xf80007ff,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			enc := NewEncoder(tc.offset, tc.size)

			// sizes
			require.Equal(t, tc.wantLeft, enc.LeftSize(), "LeftSize")
			require.Equal(t, tc.wantPre, enc.PrefixSize(), "PrefixSize")
			require.Equal(t, tc.wantRight, enc.RightSize(), "RightSize")
			require.Equal(t, 32, enc.LeftSize()+enc.PrefixSize()+enc.RightSize(),
				"sum of sizes should be 32",
			)

			// encode
			gotEnc := enc.Encode(tc.orig)
			require.Equalf(t, tc.wantEnc, uint32(gotEnc),
				"Encode(0x%08x) = 0x%08x; want 0x%08x",
				tc.orig, gotEnc, tc.wantEnc,
			)

			// prefix extractor should match the top size bits of gotEnc
			gotPre := enc.Prefix(gotEnc)
			require.Equalf(t, uint32((gotEnc>>(32-tc.size))&((1<<tc.size)-1)), gotPre,
				"Prefix(0x%08x) = 0x%08x; want 0x%08x",
				gotEnc, gotPre, gotPre,
			)

			// decode
			dec := enc.Decode(gotEnc)
			require.Equalf(t, tc.orig, dec,
				"Decode(0x%08x) = 0x%08x; want 0x%08x",
				gotEnc, dec, tc.orig,
			)
		})
	}
}

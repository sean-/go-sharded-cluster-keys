package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/sean-/go-sharded-cluster-keys/key32"
)

const maskOffset, maskSize = 11, 13

func main() {
	fmt.Printf("mask offset:\t%d\n", maskOffset)
	fmt.Printf("mask size:\t%d\n", maskSize)

	// demo values (deduped & sorted)
	values := uniqueSorted(exampleValues())

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.AlignRight)
	defer w.Flush()
	fmt.Fprintln(w, strings.Join([]string{
		"orig(dec)",
		"orig(hex)",
		"orig(bin)",

		"enc(hex)",
		"enc(bin)",

		"prefix(hex)",
		"prefix(bin)",
		"",
	}, "\t"))

	enc := key32.NewEncoder(maskOffset, maskSize)
	hexDigits := enc.PrefixHexSize()
	fmt.Printf("hex nibbles:\t%d\n", hexDigits)
	for _, v := range values {
		orig := uint32(v)
		encoded := enc.Encode(orig)
		decoded := enc.Decode(encoded)
		if decoded != orig {
			panic(fmt.Sprintf("round-trip failed: got %d, want %d", decoded, orig))
		}
		prefix := enc.Prefix(encoded)

		fmt.Fprintf(w, "%10d\t", orig)
		fmt.Fprintf(w, "%08x\t", orig)
		fmt.Fprintf(w, "%032b\t", orig)

		fmt.Fprintf(w, "%08x\t", uint32(encoded))
		fmt.Fprintf(w, "%032b\t", uint32(encoded))

		fmt.Fprintf(w, "%0*x\t", hexDigits, enc.PrefixHexPad(prefix))
		fmt.Fprintf(w, "%0*b\t", enc.PrefixSize(), prefix)
		fmt.Fprintln(w)
	}
}

func exampleValues() []uint64 {
	seq := []uint64{
		0, 1, 2, 3, 4, 127, 128, 129, 255, 256, 1023, 1024,
		(1 << maskOffset) + -1,
		(1 << maskOffset) + 0,
		(1 << maskOffset) + 1,

		(2 << maskOffset) + -1,
		(2 << maskOffset) + 0,
		(2 << maskOffset) + 1,

		(math.MaxUint32 - (1 << maskOffset)) + -1,
		(math.MaxUint32 - (1 << maskOffset)) + 0,
		(math.MaxUint32 - (1 << maskOffset)) + 1,

		math.MaxUint32 + -1,
		math.MaxUint32 + 0,
	}
	const shiftBy = maskOffset
	for i := uint64(2); i < 33; i++ {
		seq = append(seq, (i<<shiftBy)-1, i<<shiftBy, (i<<shiftBy)+1)
	}
	seq = append(seq,
		uint64(1<<31)-1,
		uint64(1<<31),
		uint64(1<<31)+1,
		uint64(1<<32)-1,
		uint64(1<<32),
	)
	for i := uint64(1 << 31); i < 33; i++ {
		seq = append(seq, (i<<shiftBy)-1, i<<shiftBy, (i<<shiftBy)+1)
	}
	return seq
}

func uniqueSorted(vals []uint64) []uint64 {
	set := make(map[uint64]struct{}, len(vals))
	for _, v := range vals {
		set[v] = struct{}{}
	}
	out := make([]uint64, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

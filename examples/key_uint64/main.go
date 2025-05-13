package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/sean-/go-sharded-cluster-keys/key64"
)

const maskOffset, maskSize = 11, 13

func main() {
	fmt.Printf("mask offset:\t%d\n", maskOffset)
	fmt.Printf("mask size:\t%d\n", maskSize)

	// demo values (deduped & sorted)
	values := uniqueSorted(exampleValues())

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, strings.Join([]string{
		"Input",
		"Decimal",
		"Hex",
		"Binary",
		"",
	}, "\t"))
	printSep(w)

	enc := key64.NewEncoder(maskOffset, maskSize)
	hexDigits := enc.PrefixHexSize()
	fmt.Printf("hex nibbles:\t%d\n", hexDigits)
	for _, v := range values {
		encoded := enc.Encode(v)
		decoded := enc.Decode(encoded)
		if decoded != v {
			panic(fmt.Sprintf("round-trip failed: got %x, want %x", decoded, v))
		}

		// original
		fmt.Fprintf(w, "%s\t", "orig")
		fmt.Fprintf(w, "%20d\t", v)
		fmt.Fprintf(w, "%016x\t", v)
		fmt.Fprintf(w, "%064b\t", v)
		fmt.Fprintln(w)

		// encoded
		fmt.Fprintf(w, "%s\t", "encoded")
		fmt.Fprintf(w, "%20d\t", encoded)
		fmt.Fprintf(w, "%016x\t", encoded)
		fmt.Fprintf(w, "%064b\t", encoded)
		fmt.Fprintln(w)

		// prefix (only maskSize bits wide)
		prefix := enc.Prefix(encoded)
		paddedPrefix := enc.PrefixHexPad(prefix)
		fmt.Fprintf(w, "%s\t", "prefix")
		fmt.Fprintf(w, "%20d\t", prefix)
		fmt.Fprintf(w, "%0*x\t", hexDigits, paddedPrefix)
		fmt.Fprintf(w, "%0*b\t", enc.PrefixSize(), paddedPrefix)
		fmt.Fprintln(w)
		printSep(w)
	}
}

func exampleValues() []uint64 {
	vals := []uint64{
		0,
		1,

		(1 << maskOffset) + -1,
		(1 << maskOffset) + 0,
		(1 << maskOffset) + 1,

		(2 << maskOffset) + -1,
		(2 << maskOffset) + 0,
		(2 << maskOffset) + 1,

		math.MaxUint32 - 1,
		math.MaxUint32 + 0,
		math.MaxUint32 + 1,

		math.MaxUint64 + -1,
		math.MaxUint64 + 0,

		0x0123456789ABCDEF + -1,
		0x0123456789ABCDEF + 0,
		0x0123456789ABCDEF + 1,
	}
	return vals
}

func printSep(w io.Writer) {
	// separator: a row of dashes in each column
	fmt.Fprintln(w, strings.Join([]string{
		"-------",              // Input
		"--------------------", // Decimal (up to 64-bit decimal)
		"----------------",     // Hex (16 chars)
		"----------------------------------------------------------------", // Binary (64 chars)
		"",
	}, "\t"))
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

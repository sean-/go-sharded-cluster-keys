# go-sharded-cluster-keys

A small Go library for prefix-shard encoding of 32- and 64-bit integers, plus
thin wrappers for UUID/ULID keys.  Useful for building time-or shard-aware keys
in distributed systems.

- [Installation](#installation)
- [Documentation](#documentation)
- [Packages](#packages)
  - `key32`
  - `key64`
  - `keyuuid`
- [Examples](#examples)

---

## Installation

```bash
go get github.com/sean-/go-sharded-cluster-keys
```

Or add to your `go.mod`:

```go
require github.com/sean-/go-sharded-cluster-keys v0.0.0
```

---

## Documentation

- 📖 pkg.go.dev:
  - key32: https://pkg.go.dev/github.com/sean-/go-sharded-cluster-keys/key32
  - key64: https://pkg.go.dev/github.com/sean-/go-sharded-cluster-keys/key64
  - keyuuid: https://pkg.go.dev/github.com/sean-/go-sharded-cluster-keys/keyuuid

---

## Packages

### `key32`

Encode, decode, and inspect 32-bit keys with a configurable “shard” segment.

```go
import "github.com/sean-/go-sharded-cluster-keys/key32"

// Create an encoder that takes 13 bits starting at bit-11,
// reverses them, and prepends them into the top 13 bits.
enc32 := key32.NewEncoder(11, 13)

orig32 := uint32(0x0000FFFF)
encoded := enc32.Encode(orig32)    // Value(uint32)
decoded := enc32.Decode(encoded)   // uint32 == orig32
prefix  := enc32.Prefix(encoded)   // the reversed-mask field
```

### `key64`

Same API as key32, but for 64-bit values.

```go
import "github.com/sean-/go-sharded-cluster-keys/key64"

enc64 := key64.NewEncoder(11, 13)

orig64 := uint64(0x0123456789ABCDEF)
encoded64 := enc64.Encode(orig64)    // Value(uint64)
decoded64 := enc64.Decode(encoded64) // uint64 == orig64
prefix64  := enc64.Prefix(encoded64)
```

### `keyuuid`

Encode, decode, and inspect 128-bit UUID/ULID values, exactly like `key32`/`key64`.

```go
import (
  "fmt"

  "github.com/google/uuid"
  "github.com/oklog/ulid/v2"
  "github.com/sean-/go-sharded-cluster-keys/keyuuid"
)

// 1) Generic UUID (identity encoder)
u := uuid.MustParse("3d813cbb-47fb-32ba-91df-831e1593ac29")
encGen := keyuuid.NewEncoder(0, 0, 0)         // identity over 128 bits
encU := encGen.Encode(u)                     // same as u
decoded := encGen.Decode(encU)               // equals u
prefix := encGen.Prefix(encU)                // zero UUID
fmt.Println("Identity:", encU, decoded, prefix)

// 2) UUIDv7 (48-bit timestamp + 4-bit shard at offset 11)
u7 := uuid.MustParse("018f14e0-8f0a-7def-91b4-f0ecb69f5f01")
enc7 := keyuuid.NewUUIDv7Encoder()           // totalBits=48, offset=11, size=4
enc7U := enc7.Encode(u7)
dec7U := enc7.Decode(enc7U)
pre7U := enc7.Prefix(enc7U)
fmt.Println("UUIDv7:", enc7U, dec7U, pre7U)

// 3) ULID (48-bit timestamp + 16-bit shard at offset 16)
ulidStr := "01ARYZ6S41TSV4RRFFQ69G5FAV"
uULID, _ := ulid.Parse(ulidStr)
var raw [16]byte
copy(raw[:], uULID[:])
// reinterpret ULID bytes as a UUID
var base uuid.UUID
copy(base[:], raw[:])
encUld := keyuuid.NewULIDEncoder()           // totalBits=48, offset=16, size=16
encULIDU := encUld.Encode(base)
decULIDU := encUld.Decode(encULIDU)
preULIDU := encUld.Prefix(encULIDU)
fmt.Println("ULID:", encULIDU, decULIDU, preULIDU)
```

---

## Examples

```go
package main

import (
  "fmt"

  "github.com/google/uuid"
  "github.com/oklog/ulid/v2"

  "github.com/sean-/go-sharded-cluster-keys/key32"
  "github.com/sean-/go-sharded-cluster-keys/key64"
  "github.com/sean-/go-sharded-cluster-keys/keyuuid"
)

func main() {
  // key32
  e32 := key32.NewEncoder(11,13)
  x32 := uint32(42)
  enc32 := e32.Encode(x32)
  fmt.Printf("32-bit: %d → %08x → %d\n", x32, uint32(enc32), e32.Decode(enc32))

  // key64
  e64 := key64.NewEncoder(8,16)
  x64 := uint64(0xDEADBEEFCAFEBABE)
  enc64 := e64.Encode(x64)
  fmt.Printf("64-bit: %x → %x → %x\n", x64, uint64(enc64), e64.Decode(enc64))

  // UUID
  u := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
  k := keyuuid.New(u)
  fmt.Println("UUID:", k.String())

  // UUIDv7
  u7 := uuid.MustParse("018f14e0-8f0a-7def-91b4-f0ecb69f5f01")
  k7, _ := keyuuid.NewFromUUIDv7(u7)
  fmt.Println("UUIDv7:", k7.String())

  // ULID
  uulid, _ := ulid.New(ulid.Now(), nil)
  var raw [16]byte
  copy(raw[:], uulid[:])
  ku, _ := keyuuid.NewFromULID(raw)
  fmt.Println("ULID → Key.UUID:", ku.UUID())
}
```

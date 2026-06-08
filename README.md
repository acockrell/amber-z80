# amber-z80

[![CI](https://github.com/acockrell/amber-z80/actions/workflows/ci.yml/badge.svg)](https://github.com/acockrell/amber-z80/actions/workflows/ci.yml)
[![ZEXALL](https://github.com/acockrell/amber-z80/actions/workflows/zexall.yml/badge.svg)](https://github.com/acockrell/amber-z80/actions/workflows/zexall.yml)

A Zilog Z80 CPU emulator in Go. Cycle-accurate, instruction-faithful, and validated against the standard test suite.

Originally extracted from the [amber](https://github.com/acockrell/amber) CP/M 3.0 emulator.

## Status

Passes **all 67 tests** of the ZEXALL Z80 instruction validation suite, including:

- Undocumented F3/F5 flag bits on every arithmetic, logical, and rotate instruction
- NMOS SCF/CCF flag quirks (behaviour depending on the previous `Q` value)
- MEMPTR (WZ) internal register tracking, observable via `BIT n,(HL)`
- All documented and undocumented DD/FD/ED/CB prefixes
- DAA across both addition and subtraction with every flag combination

Every opcode returns its real T-state count, so consumers that need cycle-accurate timing (mid-scanline interrupts, contended memory, beam-racing tricks on other Z80 platforms) can sync against `CPU.Cycles`.

## Install

```sh
go get github.com/acockrell/amber-z80
```

## Use

```go
package main

import "github.com/acockrell/amber-z80"

// Provide a Memory implementation — anything mapping uint16 to byte.
type ram [65536]byte
func (r *ram) Read(addr uint16) byte       { return r[addr] }
func (r *ram) Write(addr uint16, val byte) { r[addr] = val }

func main() {
    mem := &ram{}
    mem.Write(0x0100, 0x3E) // LD A, 0x42
    mem.Write(0x0101, 0x42)
    mem.Write(0x0102, 0x76) // HALT

    cpu := z80.New(mem)
    cpu.PC = 0x0100
    for !cpu.Halted {
        cpu.Step()
    }
    // cpu.A == 0x42
}
```

### Port I/O

```go
cpu.In  = func(port uint16) byte { /* … */ return 0 }
cpu.Out = func(port uint16, val byte) { /* … */ }
```

### Interrupts

```go
cpu.NMI()   // non-maskable, jumps to 0x0066
cpu.INT()   // maskable; honours IFF1 and IM 0/1/2
```

## What's not in scope

- A built-in memory implementation. The `Memory` interface is two methods (`Read`, `Write`); bring your own — flat 64 KB, banked, memory-mapped I/O, whatever your platform needs.
- Disassembly / debugger.
- Cycle-accurate bus contention (the CPU exposes T-states; sequencing them against external hardware is the consumer's job).

## Validation

ZEXDOC and ZEXALL are Frank Cringle's Z80 instruction exercisers, originally
written for his YAZE Z80 emulator and ported to CP/M by J.G. Harston.
They are distributed under the GNU General Public License. Source and
pre-assembled binaries are available at
[mdfs.net/Software/Z80/Exerciser](https://mdfs.net/Software/Z80/Exerciser/).
ZEXDOC tests documented behaviour; ZEXALL additionally covers undocumented
opcodes and flag bits. Both binaries are vendored under `testdata/`.

CI runs them automatically:

| Suite  | When           | Duration | Coverage                        |
|--------|----------------|----------|---------------------------------|
| ZEXDOC | every PR       | ~3 min   | all documented instructions     |
| ZEXALL | push to `main` | ~10 min  | documented + undocumented       |

To run locally:

```sh
go run ./cmd/zexall testdata/zexdoc.com   # fast
go run ./cmd/zexall testdata/zexall.com   # full
```

The runner accepts any CP/M `.com` binary with a minimal BDOS console stub
(functions 2 and 9). It exits 1 if any test reports errors.

## Development

```sh
make test         # unit tests
make test-race    # tests with -race
make lint         # golangci-lint
make cover-html   # coverage report
make bench        # benchmarks
```

CI runs `go vet`, `go test -race -coverprofile`, `golangci-lint`, and ZEXDOC/ZEXALL when Go source, module files, or workflow configs change.

## Layout

```
doc.go             package-level documentation
z80.go             CPU struct, registers, fetch/push/pop, Step, NMI, INT
flags.go           F register helpers and SZ53 lookup
opcodes.go         primary instruction table (0x00–0xFF except prefix bytes)
opcodes_cb.go      CB-prefix: rotates, shifts, BIT/SET/RES
opcodes_dd.go      DD-prefix: IX-indexed instructions
opcodes_ed.go      ED-prefix: extended ops (block moves, I/O, IM, etc.)
opcodes_fd.go      FD-prefix: IY-indexed instructions
cmd/zexall/        standalone CP/M test runner (ZEXALL, ZEXDOC, PRELIM, …)
testdata/          vendored ZEXALL and ZEXDOC binaries
```

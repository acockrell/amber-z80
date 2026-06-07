# amber-z80

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

ZEXALL is run via the parent [amber](https://github.com/acockrell/amber) emulator, which provides the CP/M BDOS console plumbing the test binary needs:

```sh
git clone https://github.com/acockrell/amber
cd amber && make zexall
```

Expect roughly 10 minutes to complete all 67 tests.

## Layout

```
z80.go         CPU struct, registers, fetch/push/pop, Step, NMI, INT
flags.go       F register helpers and SZ53 lookup
opcodes.go     primary instruction table (0x00–0xFF except prefix bytes)
opcodes_cb.go  CB-prefix: rotates, shifts, BIT/SET/RES
opcodes_dd.go  DD-prefix: IX-indexed instructions
opcodes_ed.go  ED-prefix: extended ops (block moves, I/O, IM, etc.)
opcodes_fd.go  FD-prefix: IY-indexed instructions
```

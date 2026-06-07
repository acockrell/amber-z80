// Command zexall runs a CP/M .com binary (ZEXALL, ZEXDOC, PRELIM, etc.) against
// the z80 core with a minimal BDOS console stub. Exit code is 0 on clean
// termination (warm boot / HALT), 1 on harness error.
//
//	go run ./cmd/zexall path/to/zexall.com
package main

import (
	"fmt"
	"os"
	"time"

	z80 "github.com/acockrell/amber-z80"
)

const (
	bdosEntry  = 0x0005
	wbootEntry = 0x0000
	loadAddr   = 0x0100
)

type ram [65536]byte

func (r *ram) Read(addr uint16) byte       { return r[addr] }
func (r *ram) Write(addr uint16, val byte) { r[addr] = val }

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: zexall <path-to-cpm-binary.com>")
		os.Exit(2)
	}
	path := os.Args[1]

	binary, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	if len(binary) > 65536-loadAddr {
		fmt.Fprintf(os.Stderr, "binary too large: %d bytes\n", len(binary))
		os.Exit(1)
	}

	mem := &ram{}
	copy(mem[loadAddr:], binary)

	// CP/M-style traps:
	//   0x0000: warm boot — terminate
	//   0x0005: BDOS call — handle, then RET
	mem[wbootEntry] = 0x76 // HALT
	mem[bdosEntry] = 0xC9  // RET (we service the call before the fetch below)

	c := z80.New(mem)
	c.PC = loadAddr
	c.SP = 0xF000

	start := time.Now()
	var steps uint64

	for !c.Halted {
		switch c.PC {
		case bdosEntry:
			switch c.C {
			case 2: // C_WRITE — print char in E
				os.Stdout.Write([]byte{c.E})
			case 9: // C_WRITESTR — print $-terminated string at DE
				for a := c.DE(); mem.Read(a) != '$'; a++ {
					os.Stdout.Write([]byte{mem.Read(a)})
				}
			}
		case wbootEntry:
			// Program returned to 0x0000 — let the HALT execute and exit.
		}
		c.Cycles += uint64(c.Step())
		steps++
	}

	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "\n[%d steps, %d t-states, %s, ~%.1f MHz equivalent]\n",
		steps, c.Cycles, elapsed.Round(time.Millisecond),
		float64(c.Cycles)/elapsed.Seconds()/1e6)
}

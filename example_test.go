package z80_test

import (
	"fmt"

	z80 "github.com/acockrell/amber-z80"
)

// flat64k is a minimal Memory implementation backed by a 64 KB array.
type flat64k [65536]byte

func (m *flat64k) Read(addr uint16) byte       { return m[addr] }
func (m *flat64k) Write(addr uint16, val byte) { m[addr] = val }

// Example_haltLoop shows the basic pattern for running a program to completion.
func Example_haltLoop() {
	mem := &flat64k{}
	// LD A, 0x42 ; HALT
	mem[0x0100] = 0x3E
	mem[0x0101] = 0x42
	mem[0x0102] = 0x76

	cpu := z80.New(mem)
	cpu.PC = 0x0100

	for !cpu.Halted {
		cpu.Cycles += uint64(cpu.Step())
	}

	fmt.Printf("A = 0x%02X, cycles = %d\n", cpu.A, cpu.Cycles)
	// Output:
	// A = 0x42, cycles = 11
}

// Example_portIO shows how to wire up port I/O handlers.
func Example_portIO() {
	mem := &flat64k{}
	// IN A, (0x10) ; HALT
	mem[0x0100] = 0xDB
	mem[0x0101] = 0x10
	mem[0x0102] = 0x76

	cpu := z80.New(mem)
	cpu.PC = 0x0100

	// IN A,(n) uses port = (A<<8)|n. A is 0, n is 0x10, so port = 0x0010.
	ports := map[uint16]byte{0x0010: 0xAB}
	cpu.In = func(port uint16) byte { return ports[port] }

	for !cpu.Halted {
		cpu.Cycles += uint64(cpu.Step())
	}

	fmt.Printf("A = 0x%02X\n", cpu.A)
	// Output:
	// A = 0xAB
}

// Example_nmi shows how to deliver a non-maskable interrupt.
func Example_nmi() {
	mem := &flat64k{}
	// 0x0066: NMI handler — LD A, 0xFF ; RETN
	mem[0x0066] = 0x3E
	mem[0x0067] = 0xFF
	mem[0x0068] = 0xED
	mem[0x0069] = 0x45

	// Main program: LD A, 0x00 ; HALT ; HALT
	// The second HALT is the resume point after RETN returns to 0x0103.
	mem[0x0100] = 0x3E
	mem[0x0101] = 0x00
	mem[0x0102] = 0x76
	mem[0x0103] = 0x76

	cpu := z80.New(mem)
	cpu.PC = 0x0100
	cpu.SP = 0xF000

	// Run until halt, then deliver NMI, then run again.
	for !cpu.Halted {
		cpu.Cycles += uint64(cpu.Step())
	}
	cpu.NMI()
	for !cpu.Halted {
		cpu.Cycles += uint64(cpu.Step())
	}

	fmt.Printf("A = 0x%02X\n", cpu.A)
	// Output:
	// A = 0xFF
}

// ExampleCPU_Step shows how to use T-state counts for cycle-accurate timing.
func ExampleCPU_Step() {
	mem := &flat64k{}
	// NOP (4 T-states) ; LD BC, nn (10 T-states) ; HALT
	mem[0x0100] = 0x00
	mem[0x0101] = 0x01
	mem[0x0102] = 0x34
	mem[0x0103] = 0x12
	mem[0x0104] = 0x76

	cpu := z80.New(mem)
	cpu.PC = 0x0100

	for !cpu.Halted {
		tstates := cpu.Step()
		cpu.Cycles += uint64(tstates)
		fmt.Printf("PC after: 0x%04X  T-states: %d\n", cpu.PC, tstates)
	}
	// Output:
	// PC after: 0x0101  T-states: 4
	// PC after: 0x0104  T-states: 10
	// PC after: 0x0105  T-states: 4
}

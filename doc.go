// Package z80 implements a cycle-accurate Zilog Z80 CPU emulator.
//
// # Overview
//
// The package exposes a single [CPU] type. Callers provide a [Memory]
// implementation — anything that maps a uint16 address to a byte — and the CPU
// does the rest: fetch, decode, execute, update flags, count T-states.
//
// The emulator is validated against Frank Cringle's ZEXALL instruction
// exerciser, which covers all 67 documented and undocumented instruction
// groups including F3/F5 flag bits, NMOS SCF/CCF quirks, MEMPTR (WZ)
// tracking, and all DD/FD/ED/CB prefix combinations.
//
// # Getting started
//
//	type ram [65536]byte
//	func (r *ram) Read(addr uint16) byte       { return r[addr] }
//	func (r *ram) Write(addr uint16, val byte) { r[addr] = val }
//
//	mem := &ram{}
//	cpu := z80.New(mem)
//	cpu.PC = 0x0100
//	for !cpu.Halted {
//	    cpu.Step()
//	}
//
// # Timing
//
// [CPU.Step] returns the number of T-states consumed by the instruction.
// [CPU.Cycles] accumulates the total T-state count across all calls to [CPU.Step]
// and [CPU.Run]. At 3.5 MHz (original ZX Spectrum clock), one T-state is
// ~286 ns; at 4 MHz (CP/M systems), ~250 ns.
//
// # Port I/O
//
// Assign [CPU.In] and [CPU.Out] before running:
//
//	cpu.In  = func(port uint16) byte { return myBus.Read(port) }
//	cpu.Out = func(port uint16, val byte) { myBus.Write(port, val) }
//
// Both default to no-ops (In returns 0xFF, Out discards the value).
//
// # Interrupts
//
// Call [CPU.NMI] or [CPU.INT] from the host at the appropriate T-state
// boundary. [CPU.INT] respects IFF1 and the current interrupt mode (IM 0/1/2).
// [CPU.NMI] is always accepted. Both clear the halted state.
//
// # What is not included
//
// The package ships no [Memory] implementation — the interface is two methods
// and callers bring their own (flat RAM, banked ROM/RAM, memory-mapped I/O).
// Disassembly, a debugger interface, and bus-contention modelling are also out
// of scope; the T-state counts from [CPU.Step] give callers everything they
// need to implement contention externally.
package z80

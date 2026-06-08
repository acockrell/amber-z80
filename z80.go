package z80

// Memory is the address space the CPU executes against. Implementations must
// behave as a flat 64 KB byte array indexed by uint16 — the CPU does not care
// what backs it (linear RAM, banked RAM, memory-mapped I/O, etc.).
type Memory interface {
	Read(addr uint16) byte
	Write(addr uint16, val byte)
}

// CPU is the Zilog Z80 processor.
type CPU struct {
	// Main register set
	A, F byte
	B, C byte
	D, E byte
	H, L byte

	// Alternate register set
	A_, F_ byte
	B_, C_ byte
	D_, E_ byte
	H_, L_ byte

	// Index registers
	IX, IY uint16

	// Stack pointer and program counter
	SP, PC uint16

	// Interrupt vector / memory refresh
	I, R byte

	// Internal memory pointer (used for undocumented flag bits)
	WZ uint16

	// Interrupt flip-flops and mode
	IFF1, IFF2 bool
	IM         byte // 0, 1, or 2

	// Halted state
	Halted bool

	// Pending EI: delay interrupt enable by one instruction
	pendingEI bool

	// Memory
	Mem Memory

	// I/O port handlers — set by BIOS/BDOS layer
	In  func(port uint16) byte
	Out func(port uint16, val byte)

	// Cycles executed (for timing)
	Cycles uint64
}

// New returns a CPU connected to mem with I/O ports defaulting to no-ops.
// Set [CPU.In], [CPU.Out], and [CPU.PC] before calling [CPU.Step].
func New(mem Memory) *CPU {
	c := &CPU{Mem: mem}
	c.In = func(port uint16) byte { return 0xFF }
	c.Out = func(port uint16, val byte) {}
	return c
}

// AF returns the AF register pair (A in the high byte, F in the low byte).
func (c *CPU) AF() uint16 { return uint16(c.A)<<8 | uint16(c.F) }

// SetAF sets the AF register pair.
func (c *CPU) SetAF(v uint16) { c.A = byte(v >> 8); c.F = byte(v) }

// BC returns the BC register pair.
func (c *CPU) BC() uint16 { return uint16(c.B)<<8 | uint16(c.C) }

// SetBC sets the BC register pair.
func (c *CPU) SetBC(v uint16) { c.B = byte(v >> 8); c.C = byte(v) }

// DE returns the DE register pair.
func (c *CPU) DE() uint16 { return uint16(c.D)<<8 | uint16(c.E) }

// SetDE sets the DE register pair.
func (c *CPU) SetDE(v uint16) { c.D = byte(v >> 8); c.E = byte(v) }

// HL returns the HL register pair.
func (c *CPU) HL() uint16 { return uint16(c.H)<<8 | uint16(c.L) }

// SetHL sets the HL register pair.
func (c *CPU) SetHL(v uint16) { c.H = byte(v >> 8); c.L = byte(v) }

// fetch reads the byte at PC and increments PC.
func (c *CPU) fetch() byte {
	v := c.Mem.Read(c.PC)
	c.PC++
	c.R = (c.R & 0x80) | ((c.R + 1) & 0x7F)
	return v
}

// fetch16 reads a little-endian word at PC and increments PC by 2.
func (c *CPU) fetch16() uint16 {
	lo := c.fetch()
	hi := c.fetch()
	return uint16(hi)<<8 | uint16(lo)
}

// read16 reads a little-endian word from memory.
func (c *CPU) read16(addr uint16) uint16 {
	return uint16(c.Mem.Read(addr)) | uint16(c.Mem.Read(addr+1))<<8
}

// write16 writes a little-endian word to memory.
func (c *CPU) write16(addr uint16, v uint16) {
	c.Mem.Write(addr, byte(v))
	c.Mem.Write(addr+1, byte(v>>8))
}

// push pushes a 16-bit value onto the stack.
func (c *CPU) push(v uint16) {
	c.SP -= 2
	c.write16(c.SP, v)
}

// pop pops a 16-bit value from the stack.
func (c *CPU) pop() uint16 {
	v := c.read16(c.SP)
	c.SP += 2
	return v
}

// Step executes one instruction and returns the number of T-states consumed.
// If the CPU is halted it returns 4 without advancing PC (the real Z80
// repeatedly executes NOP internally while halted; call [CPU.NMI] or
// [CPU.INT] to resume). Step does not update [CPU.Cycles]; callers that need
// a running total should add the return value themselves, or use [CPU.Run].
func (c *CPU) Step() int {
	if c.Halted {
		return 4
	}

	if c.pendingEI {
		c.pendingEI = false
		c.IFF1 = true
		c.IFF2 = true
	}

	op := c.fetch()
	return c.execOp(op)
}

// Run executes instructions until the halt channel receives a signal.
// It accumulates T-states into [CPU.Cycles]. For finer-grained control
// (mid-scanline interrupts, contention) use [CPU.Step] directly.
func (c *CPU) Run(halt <-chan struct{}) {
	for {
		select {
		case <-halt:
			return
		default:
			c.Cycles += uint64(c.Step())
		}
	}
}

// Reset resets the CPU to power-on state: PC=0, SP=0xFFFF, A=0xFF, F=0xFF,
// interrupts disabled, IM 0, Halted false, Cycles 0.
func (c *CPU) Reset() {
	c.PC = 0
	c.SP = 0xFFFF
	c.A = 0xFF
	c.F = 0xFF
	c.IFF1 = false
	c.IFF2 = false
	c.IM = 0
	c.Halted = false
	c.I = 0
	c.R = 0
	c.Cycles = 0
}

// NMI triggers a non-maskable interrupt. IFF1 is saved to IFF2, IFF1 is
// cleared, the return address is pushed, and PC jumps to 0x0066. If the CPU
// is halted it resumes. Costs 11 T-states.
func (c *CPU) NMI() {
	c.Halted = false
	c.IFF2 = c.IFF1
	c.IFF1 = false
	c.push(c.PC)
	c.PC = 0x0066
	c.Cycles += 11
}

// INT triggers a maskable interrupt. It is ignored if IFF1 is false.
// Both IFF1 and IFF2 are cleared. Behaviour by mode:
//   - IM 0/1: push PC, jump to 0x0038 (13 T-states)
//   - IM 2: push PC, read vector from (I<<8 | 0xFF), jump there (19 T-states)
//
// If the CPU is halted it resumes before the interrupt is serviced.
func (c *CPU) INT() {
	if !c.IFF1 {
		return
	}
	c.Halted = false
	c.IFF1 = false
	c.IFF2 = false
	switch c.IM {
	case 0, 1:
		c.push(c.PC)
		c.PC = 0x0038
		c.Cycles += 13
	case 2:
		c.push(c.PC)
		vec := uint16(c.I)<<8 | 0xFF
		c.PC = c.read16(vec)
		c.Cycles += 19
	}
}

// readReg returns the value of an 8-bit register by index (0=B,1=C,...,6=(HL),7=A).
func (c *CPU) readReg(r byte) byte {
	switch r {
	case 0:
		return c.B
	case 1:
		return c.C
	case 2:
		return c.D
	case 3:
		return c.E
	case 4:
		return c.H
	case 5:
		return c.L
	case 6:
		return c.Mem.Read(c.HL())
	case 7:
		return c.A
	}
	return 0
}

// writeReg writes to an 8-bit register by index.
func (c *CPU) writeReg(r, val byte) {
	switch r {
	case 0:
		c.B = val
	case 1:
		c.C = val
	case 2:
		c.D = val
	case 3:
		c.E = val
	case 4:
		c.H = val
	case 5:
		c.L = val
	case 6:
		c.Mem.Write(c.HL(), val)
	case 7:
		c.A = val
	}
}

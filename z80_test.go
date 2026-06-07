package z80

import "testing"

// ram is a trivial 64 KB linear memory used by the test suite. The library
// itself ships no Memory implementation — consumers (or tests like these)
// provide their own.
type ram [65536]byte

func (r *ram) Read(addr uint16) byte       { return r[addr] }
func (r *ram) Write(addr uint16, val byte) { r[addr] = val }
func (r *ram) load(addr uint16, data []byte) {
	copy(r[addr:], data)
}

func newRAM() *ram { return &ram{} }

func newCPU(prog []byte) *CPU {
	mem := newRAM()
	mem.load(0x0100, prog)
	c := New(mem)
	c.PC = 0x0100
	return c
}

func TestLDImmediate(t *testing.T) {
	c := newCPU([]byte{0x3E, 0x42}) // LD A, 0x42
	c.Step()
	if c.A != 0x42 {
		t.Fatalf("A = %02X, want 0x42", c.A)
	}
}

func TestAddSub(t *testing.T) {
	c := newCPU([]byte{
		0x3E, 0x10, // LD A, 0x10
		0xC6, 0x05, // ADD A, 5
		0xD6, 0x03, // SUB 3
	})
	c.Step()
	c.Step()
	c.Step()
	if c.A != 0x12 {
		t.Fatalf("A = %02X, want 0x12", c.A)
	}
}

func TestFlags(t *testing.T) {
	c := newCPU([]byte{
		0x3E, 0x01, // LD A, 1
		0xD6, 0x01, // SUB 1 → A=0, Z set
	})
	c.Step()
	c.Step()
	if c.A != 0 {
		t.Fatalf("A = %02X, want 0", c.A)
	}
	if !c.flagZ() {
		t.Fatal("Z flag not set after SUB to zero")
	}
}

func TestJR(t *testing.T) {
	c := newCPU([]byte{
		0x3E, 0x00, // LD A, 0
		0x18, 0x02, // JR +2 (skip next 2 bytes)
		0x3E, 0xFF, // LD A, 0xFF  ← should be skipped
		0x3E, 0x55, // LD A, 0x55  ← should execute
	})
	for i := 0; i < 3; i++ {
		c.Step()
	}
	if c.A != 0x55 {
		t.Fatalf("A = %02X, want 0x55", c.A)
	}
}

func TestCALLRET(t *testing.T) {
	prog := make([]byte, 32)
	// CALL 0x010A
	prog[0] = 0xCD
	prog[1] = 0x0A
	prog[2] = 0x01
	// LD A, 0xBE
	prog[3] = 0x3E
	prog[4] = 0xBE
	// HALT
	prog[5] = 0x76
	// Subroutine at 0x010A: LD A, 0xEF; RET
	prog[0x0A] = 0x3E
	prog[0x0B] = 0xEF
	prog[0x0C] = 0xC9 // RET

	c := newCPU(prog)
	c.SP = 0xF000
	c.Step() // CALL
	c.Step() // LD A, 0xEF
	c.Step() // RET
	c.Step() // LD A, 0xBE
	if c.A != 0xBE {
		t.Fatalf("A = %02X after CALL/RET sequence, want 0xBE", c.A)
	}
}

func TestLDIR(t *testing.T) {
	mem := newRAM()
	src := []byte{0x01, 0x02, 0x03, 0x04}
	mem.load(0x2000, src)
	// LDIR
	mem.load(0x0100, []byte{0xED, 0xB0, 0x76}) // LDIR; HALT
	c := New(mem)
	c.PC = 0x0100
	c.SP = 0xF000
	c.SetHL(0x2000)
	c.SetDE(0x3000)
	c.SetBC(4)
	// LDIR re-executes itself (PC reset) until BC=0; one Step per byte.
	for c.BC() != 0 {
		c.Step()
	}
	for i, b := range src {
		got := mem.Read(uint16(0x3000 + i))
		if got != b {
			t.Fatalf("dest[%d] = %02X, want %02X", i, got, b)
		}
	}
}

func TestPushPop(t *testing.T) {
	c := newCPU([]byte{
		0x01, 0x34, 0x12, // LD BC, 0x1234
		0xC5,             // PUSH BC
		0xD1,             // POP DE
	})
	c.SP = 0xF000
	c.Step()
	c.Step()
	c.Step()
	if c.DE() != 0x1234 {
		t.Fatalf("DE = %04X, want 0x1234", c.DE())
	}
}

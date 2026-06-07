package z80

import "testing"

// SLL r — undocumented Z80 instruction: shift left logical, LSB := 1.
func TestSLL_AllRegs(t *testing.T) {
	cases := []struct {
		name string
		op   byte
		set  func(*CPU)
		get  func(*CPU) byte
	}{
		{"B", 0x30, func(c *CPU) { c.B = 0x55 }, func(c *CPU) byte { return c.B }},
		{"C", 0x31, func(c *CPU) { c.C = 0x55 }, func(c *CPU) byte { return c.C }},
		{"D", 0x32, func(c *CPU) { c.D = 0x55 }, func(c *CPU) byte { return c.D }},
		{"E", 0x33, func(c *CPU) { c.E = 0x55 }, func(c *CPU) byte { return c.E }},
		{"H", 0x34, func(c *CPU) { c.H = 0x55 }, func(c *CPU) byte { return c.H }},
		{"L", 0x35, func(c *CPU) { c.L = 0x55 }, func(c *CPU) byte { return c.L }},
		{"A", 0x37, func(c *CPU) { c.A = 0x55 }, func(c *CPU) byte { return c.A }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mem := newRAM()
			mem.load(0x0100, []byte{0xCB, tc.op})
			c := New(mem)
			c.PC = 0x0100
			tc.set(c)
			c.Step()
			got := tc.get(c)
			// 0x55 << 1 | 1 = 0xAB
			if got != 0xAB {
				t.Fatalf("SLL %s: got %02X, want 0xAB", tc.name, got)
			}
		})
	}
}

func TestSLL_CarryOut(t *testing.T) {
	mem := newRAM()
	mem.load(0x0100, []byte{0xCB, 0x37}) // SLL A
	c := New(mem)
	c.PC = 0x0100
	c.A = 0x80
	c.Step()
	if c.A != 0x01 {
		t.Fatalf("SLL A: got %02X, want 0x01", c.A)
	}
	if !c.flagC() {
		t.Fatal("SLL must set carry from high bit")
	}
}

// IXH/IXL (undocumented): LD IXH, n / LD IXL, n via DD 26 nn / DD 2E nn.
func TestLD_IXH_IXL_Immediate(t *testing.T) {
	mem := newRAM()
	mem.load(0x0100, []byte{
		0xDD, 0x26, 0xAB, // LD IXH, 0xAB
		0xDD, 0x2E, 0xCD, // LD IXL, 0xCD
	})
	c := New(mem)
	c.PC = 0x0100
	c.Step()
	c.Step()
	if c.IX != 0xABCD {
		t.Fatalf("IX = %04X, want 0xABCD", c.IX)
	}
}

func TestLD_IYH_IYL_Immediate(t *testing.T) {
	mem := newRAM()
	mem.load(0x0100, []byte{
		0xFD, 0x26, 0x12,
		0xFD, 0x2E, 0x34,
	})
	c := New(mem)
	c.PC = 0x0100
	c.Step()
	c.Step()
	if c.IY != 0x1234 {
		t.Fatalf("IY = %04X, want 0x1234", c.IY)
	}
}

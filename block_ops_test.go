package z80

import "testing"

// Helper: run one ED-prefixed block op (single, non-repeating variant).
func stepED(mem *ram, op byte, setup func(*CPU)) *CPU {
	mem.load(0x0100, []byte{0xED, op})
	c := New(mem)
	c.PC = 0x0100
	setup(c)
	c.Step()
	return c
}

func TestLDI_DecrementsBC_AdvancesPointers(t *testing.T) {
	mem := newRAM()
	mem.Write(0x2000, 0xAB)
	c := stepED(mem, 0xA0, func(c *CPU) {
		c.SetHL(0x2000)
		c.SetDE(0x3000)
		c.SetBC(2)
	})
	if c.HL() != 0x2001 || c.DE() != 0x3001 {
		t.Fatalf("HL=%04X DE=%04X", c.HL(), c.DE())
	}
	if c.BC() != 1 {
		t.Fatalf("BC = %d, want 1", c.BC())
	}
	if mem.Read(0x3000) != 0xAB {
		t.Fatal("byte not transferred")
	}
	if !c.flagPV() {
		t.Fatal("P/V must be set while BC != 0 after LDI")
	}
}

func TestLDI_PV_ClearedWhenBCReachesZero(t *testing.T) {
	mem := newRAM()
	mem.Write(0x2000, 0x01)
	c := stepED(mem, 0xA0, func(c *CPU) {
		c.SetHL(0x2000)
		c.SetDE(0x3000)
		c.SetBC(1)
	})
	if c.BC() != 0 {
		t.Fatalf("BC = %d, want 0", c.BC())
	}
	if c.flagPV() {
		t.Fatal("P/V must be cleared when BC reaches 0 after LDI")
	}
}

func TestLDD_DecrementsPointers(t *testing.T) {
	mem := newRAM()
	mem.Write(0x2005, 0x5A)
	c := stepED(mem, 0xA8, func(c *CPU) {
		c.SetHL(0x2005)
		c.SetDE(0x3005)
		c.SetBC(3)
	})
	if c.HL() != 0x2004 || c.DE() != 0x3004 {
		t.Fatalf("HL=%04X DE=%04X (LDD must decrement)", c.HL(), c.DE())
	}
	if mem.Read(0x3005) != 0x5A {
		t.Fatal("byte not transferred by LDD")
	}
}

func TestLDDR_CopiesBackward(t *testing.T) {
	mem := newRAM()
	src := []byte{0x11, 0x22, 0x33, 0x44}
	mem.load(0x2000, src)
	mem.load(0x0100, []byte{0xED, 0xB8})
	c := New(mem)
	c.PC = 0x0100
	c.SetHL(0x2003)
	c.SetDE(0x3003)
	c.SetBC(4)
	for c.BC() != 0 {
		c.Step()
	}
	for i, b := range src {
		if got := mem.Read(uint16(0x3000 + i)); got != b {
			t.Fatalf("dest[%d] = %02X, want %02X", i, got, b)
		}
	}
}

func TestCPI_MatchSetsZ(t *testing.T) {
	mem := newRAM()
	mem.Write(0x2000, 0x42)
	c := stepED(mem, 0xA1, func(c *CPU) {
		c.A = 0x42
		c.SetHL(0x2000)
		c.SetBC(5)
	})
	if !c.flagZ() {
		t.Fatal("CPI: Z must be set on match")
	}
	if c.HL() != 0x2001 {
		t.Fatalf("HL = %04X, want 0x2001", c.HL())
	}
	if c.BC() != 4 {
		t.Fatalf("BC = %d, want 4", c.BC())
	}
}

func TestCPIR_StopsOnMatch(t *testing.T) {
	mem := newRAM()
	mem.load(0x2000, []byte{0x01, 0x02, 0x42, 0x04})
	mem.load(0x0100, []byte{0xED, 0xB1})
	c := New(mem)
	c.PC = 0x0100
	c.A = 0x42
	c.SetHL(0x2000)
	c.SetBC(4)

	for i := 0; i < 10; i++ {
		c.Step()
		if c.PC == 0x0102 {
			break
		}
	}
	if c.PC != 0x0102 {
		t.Fatalf("CPIR did not terminate at PC after op; PC=%04X", c.PC)
	}
	if !c.flagZ() {
		t.Fatal("CPIR: Z must be set on match")
	}
	if c.HL() != 0x2003 {
		t.Fatalf("HL = %04X, want 0x2003 (past matched byte)", c.HL())
	}
}

func TestCPIR_StopsOnBCZero(t *testing.T) {
	mem := newRAM()
	mem.load(0x2000, []byte{0x01, 0x02, 0x03})
	mem.load(0x0100, []byte{0xED, 0xB1})
	c := New(mem)
	c.PC = 0x0100
	c.A = 0xFF
	c.SetHL(0x2000)
	c.SetBC(3)

	for i := 0; i < 10 && c.PC != 0x0102; i++ {
		c.Step()
	}
	if c.PC != 0x0102 {
		t.Fatalf("CPIR did not terminate; PC=%04X", c.PC)
	}
	if c.flagZ() {
		t.Fatal("CPIR: Z must NOT be set when no match found")
	}
	if c.BC() != 0 {
		t.Fatalf("BC = %d, want 0", c.BC())
	}
}

package z80

import "testing"

func TestNMI_VectorAndIFF(t *testing.T) {
	mem := newRAM()
	c := New(mem)
	c.PC = 0x1234
	c.SP = 0xF000
	c.IFF1 = true
	c.IFF2 = false

	c.NMI()

	if c.PC != 0x0066 {
		t.Fatalf("PC = %04X, want 0x0066", c.PC)
	}
	if c.IFF1 {
		t.Fatal("IFF1 must be cleared on NMI")
	}
	if !c.IFF2 {
		t.Fatal("IFF2 must hold prior IFF1 (true) after NMI")
	}
	if mem.Read(0xEFFE) != 0x34 || mem.Read(0xEFFF) != 0x12 {
		t.Fatalf("return addr on stack wrong: %02X %02X", mem.Read(0xEFFE), mem.Read(0xEFFF))
	}
	if c.SP != 0xEFFE {
		t.Fatalf("SP = %04X, want 0xEFFE", c.SP)
	}
}

func TestNMI_ExitsHalt(t *testing.T) {
	c := New(newRAM())
	c.Halted = true
	c.SP = 0xF000
	c.NMI()
	if c.Halted {
		t.Fatal("NMI must clear Halted")
	}
}

func TestRETN_RestoresIFF1(t *testing.T) {
	mem := newRAM()
	// RETN
	mem.load(0x0100, []byte{0xED, 0x45})
	c := New(mem)
	c.PC = 0x0100
	c.SP = 0xEFFE
	mem.Write(0xEFFE, 0x00)
	mem.Write(0xEFFF, 0x20)
	c.IFF1 = false
	c.IFF2 = true

	c.Step()

	if !c.IFF1 {
		t.Fatal("RETN must restore IFF1 from IFF2")
	}
	if c.PC != 0x2000 {
		t.Fatalf("PC = %04X, want 0x2000", c.PC)
	}
}

func TestINT_IgnoredWhenDisabled(t *testing.T) {
	c := New(newRAM())
	c.PC = 0x1234
	c.SP = 0xF000
	c.IFF1 = false
	c.INT()
	if c.PC != 0x1234 {
		t.Fatal("INT must be ignored when IFF1=0")
	}
}

func TestINT_IM1_Vector(t *testing.T) {
	c := New(newRAM())
	c.PC = 0x1234
	c.SP = 0xF000
	c.IFF1 = true
	c.IM = 1
	c.INT()
	if c.PC != 0x0038 {
		t.Fatalf("IM 1 vector PC = %04X, want 0x0038", c.PC)
	}
	if c.IFF1 || c.IFF2 {
		t.Fatal("INT must clear both IFF1 and IFF2")
	}
}

func TestINT_IM2_Vector(t *testing.T) {
	mem := newRAM()
	mem.Write(0x40FF, 0x78)
	mem.Write(0x4100, 0x56)
	c := New(mem)
	c.PC = 0x1234
	c.SP = 0xF000
	c.I = 0x40
	c.IFF1 = true
	c.IM = 2

	c.INT()

	if c.PC != 0x5678 {
		t.Fatalf("IM 2 PC = %04X, want 0x5678 (vector from (I<<8|FF))", c.PC)
	}
}

func TestINT_ExitsHalt(t *testing.T) {
	c := New(newRAM())
	c.Halted = true
	c.SP = 0xF000
	c.IFF1 = true
	c.IM = 1
	c.INT()
	if c.Halted {
		t.Fatal("INT must clear Halted")
	}
}

// EI delay: the instruction immediately following EI must complete with
// IFF1 still false from the CPU's perspective when an INT arrives between
// the EI and the next Step. The pendingEI flag is consumed at the start of
// the next Step, setting IFF1=true.
func TestEI_DelayedByOneInstruction(t *testing.T) {
	mem := newRAM()
	// EI; NOP
	mem.load(0x0100, []byte{0xFB, 0x00})
	c := New(mem)
	c.PC = 0x0100
	c.SP = 0xF000

	c.Step() // EI
	if c.IFF1 {
		t.Fatal("IFF1 must NOT be set immediately after EI executes")
	}
	if !c.pendingEI {
		t.Fatal("pendingEI must be set after EI")
	}

	// INT arriving here is ignored — IFF1 still false.
	c.INT()
	if c.PC == 0x0038 {
		t.Fatal("INT taken before EI delay elapsed")
	}

	c.Step() // NOP — pendingEI consumed, IFF1 now true
	if !c.IFF1 {
		t.Fatal("IFF1 must be set after the instruction following EI")
	}
}

func TestDI_ClearsIFF(t *testing.T) {
	mem := newRAM()
	mem.load(0x0100, []byte{0xF3}) // DI
	c := New(mem)
	c.PC = 0x0100
	c.IFF1 = true
	c.IFF2 = true
	c.Step()
	if c.IFF1 || c.IFF2 {
		t.Fatal("DI must clear both IFF1 and IFF2")
	}
}

func TestIM_Modes(t *testing.T) {
	cases := []struct {
		op   byte
		want byte
	}{
		{0x46, 0}, // IM 0
		{0x56, 1}, // IM 1
		{0x5E, 2}, // IM 2
	}
	for _, tc := range cases {
		mem := newRAM()
		mem.load(0x0100, []byte{0xED, tc.op})
		c := New(mem)
		c.PC = 0x0100
		c.Step()
		if c.IM != tc.want {
			t.Fatalf("ED %02X: IM=%d, want %d", tc.op, c.IM, tc.want)
		}
	}
}

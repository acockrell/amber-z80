package z80

import "testing"

func newCPUIO(prog []byte, in func(uint16) byte, out func(uint16, byte)) *CPU {
	mem := newRAM()
	mem.load(0x0100, prog)
	c := New(mem)
	c.PC = 0x0100
	if in != nil {
		c.In = in
	}
	if out != nil {
		c.Out = out
	}
	return c
}

func TestIN_A_n(t *testing.T) {
	var seenPort uint16
	c := newCPUIO([]byte{0xDB, 0x42}, func(p uint16) byte {
		seenPort = p
		return 0x99
	}, nil)
	c.A = 0x12
	c.Step()
	// Port = (A << 8) | n
	if seenPort != 0x1242 {
		t.Fatalf("port = %04X, want 0x1242 (A:n)", seenPort)
	}
	if c.A != 0x99 {
		t.Fatalf("A = %02X, want 0x99", c.A)
	}
}

func TestOUT_n_A(t *testing.T) {
	var port uint16
	var val byte
	c := newCPUIO([]byte{0xD3, 0x42}, nil, func(p uint16, v byte) {
		port = p
		val = v
	})
	c.A = 0x77
	c.Step()
	if port != 0x7742 {
		t.Fatalf("port = %04X, want 0x7742", port)
	}
	if val != 0x77 {
		t.Fatalf("val = %02X, want 0x77", val)
	}
}

func TestIN_r_C_SetsFlags(t *testing.T) {
	// ED 40 = IN B,(C). Input = 0x80 → S set, Z clear, P clear, H clear, N clear.
	c := newCPUIO([]byte{0xED, 0x40}, func(uint16) byte { return 0x80 }, nil)
	c.SetBC(0x1234)
	c.Step()
	if c.B != 0x80 {
		t.Fatalf("B = %02X, want 0x80", c.B)
	}
	if !c.flagS() {
		t.Fatal("S must be set for 0x80 input")
	}
	if c.flagZ() {
		t.Fatal("Z must be cleared for non-zero input")
	}
	if c.flagN() {
		t.Fatal("N must be cleared by IN r,(C)")
	}
	if c.flagH() {
		t.Fatal("H must be cleared by IN r,(C)")
	}
}

func TestIN_r_C_ZeroSetsZ(t *testing.T) {
	c := newCPUIO([]byte{0xED, 0x40}, func(uint16) byte { return 0 }, nil)
	c.SetBC(0x0001)
	c.Step()
	if !c.flagZ() {
		t.Fatal("Z must be set for zero input")
	}
	if !c.flagPV() {
		t.Fatal("P/V (parity) must be set for 0 (even number of 1 bits)")
	}
}

package z80

import "testing"

// DAA reference values from Zilog manual / ZEX. These cover the documented
// post-arith adjustments without re-deriving the full truth table.
func TestDAA_PostAdd(t *testing.T) {
	cases := []struct {
		a     byte
		h, c  bool
		wantA byte
		wantC bool
	}{
		// 0x09 + 0x01 = 0x0A → DAA → 0x10
		{0x0A, false, false, 0x10, false},
		// 0x15 + 0x27 = 0x3C → DAA → 0x42 (BCD 42)
		{0x3C, false, false, 0x42, false},
		// 0x90 + 0x80 = 0x10 with C=1 → DAA → 0x70 carry
		{0x10, false, true, 0x70, true},
		// 0x99 + 0x01 = 0x9A → DAA → 0x00 carry
		{0x9A, false, false, 0x00, true},
	}
	for _, tc := range cases {
		mem := newRAM()
		mem.load(0x0100, []byte{0x27}) // DAA
		c := New(mem)
		c.PC = 0x0100
		c.A = tc.a
		c.setFlag(FlagN, false) // post-add
		c.setFlag(FlagH, tc.h)
		c.setFlag(FlagC, tc.c)
		c.Step()
		if c.A != tc.wantA {
			t.Errorf("DAA(A=%02X H=%v C=%v) → A=%02X, want %02X", tc.a, tc.h, tc.c, c.A, tc.wantA)
		}
		if c.flagC() != tc.wantC {
			t.Errorf("DAA(A=%02X) → C=%v, want %v", tc.a, c.flagC(), tc.wantC)
		}
	}
}

func TestDAA_PostSub(t *testing.T) {
	cases := []struct {
		a     byte
		h, c  bool
		wantA byte
		wantC bool
	}{
		// 0x42 - 0x15 = 0x2D, H=1 → DAA → 0x27
		{0x2D, true, false, 0x27, false},
		// 0x05 - 0x21 = 0xE4, C=1 → DAA → 0x84 carry
		{0xE4, false, true, 0x84, true},
	}
	for _, tc := range cases {
		mem := newRAM()
		mem.load(0x0100, []byte{0x27})
		c := New(mem)
		c.PC = 0x0100
		c.A = tc.a
		c.setFlag(FlagN, true) // post-sub
		c.setFlag(FlagH, tc.h)
		c.setFlag(FlagC, tc.c)
		c.Step()
		if c.A != tc.wantA {
			t.Errorf("DAA-sub(A=%02X H=%v C=%v) → A=%02X, want %02X", tc.a, tc.h, tc.c, c.A, tc.wantA)
		}
		if c.flagC() != tc.wantC {
			t.Errorf("DAA-sub(A=%02X) → C=%v, want %v", tc.a, c.flagC(), tc.wantC)
		}
		if !c.flagN() {
			t.Error("DAA must preserve N flag")
		}
	}
}

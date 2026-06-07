package z80

import "testing"

// loopCPU loads prog at 0x0100 and wraps PC back to 0x0100 every wrap steps.
// Used by benchmarks to keep running a short kernel without falling off RAM.
func benchCPU(prog []byte) *CPU {
	mem := newRAM()
	mem.load(0x0100, prog)
	c := New(mem)
	c.PC = 0x0100
	c.SP = 0xF000
	return c
}

func BenchmarkStep_NOP(b *testing.B) {
	prog := make([]byte, 256)
	// fill with NOPs, last byte = JP 0x0100
	for i := range prog[:253] {
		prog[i] = 0x00
	}
	prog[253] = 0xC3 // JP nn
	prog[254] = 0x00
	prog[255] = 0x01
	c := benchCPU(prog)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Step()
	}
	b.ReportMetric(float64(c.Cycles), "cycles")
}

func BenchmarkStep_Mixed(b *testing.B) {
	// LD A,n / ADD A,n / LD HL,nn / LD (HL),A / JP 0x0100
	prog := []byte{
		0x3E, 0x12, // LD A,0x12
		0xC6, 0x34, // ADD A,0x34
		0x21, 0x00, 0x40, // LD HL,0x4000
		0x77,             // LD (HL),A
		0xC3, 0x00, 0x01, // JP 0x0100
	}
	c := benchCPU(prog)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Cycles += uint64(c.Step())
	}
	b.ReportMetric(float64(c.Cycles)/b.Elapsed().Seconds(), "tstates/s")
}

func BenchmarkStep_PrefixCB(b *testing.B) {
	// CB 27 = SLA A ; loop
	prog := []byte{
		0xCB, 0x27,
		0xC3, 0x00, 0x01, // JP 0x0100
	}
	c := benchCPU(prog)
	c.A = 0x55
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Step()
	}
}

func BenchmarkStep_PrefixDDCB(b *testing.B) {
	// DD CB 00 06 = RLC (IX+0)
	prog := []byte{
		0xDD, 0xCB, 0x00, 0x06,
		0xC3, 0x00, 0x01,
	}
	c := benchCPU(prog)
	c.IX = 0x4000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Step()
	}
}

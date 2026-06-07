package z80

// execED handles the 0xED prefix (extended instructions).
func (c *CPU) execED() int {
	op := c.fetch()
	switch op {

	// IN r, (C)
	case 0x40, 0x48, 0x50, 0x58, 0x60, 0x68, 0x70, 0x78:
		port := uint16(c.B)<<8 | uint16(c.C)
		v := c.In(port)
		r := (op >> 3) & 7
		if r != 6 { // 0x70 reads but discards (IN F, (C))
			c.writeReg(r, v)
		}
		c.F = (c.F & FlagC) | flagsFromByte(v)
		return 12

	// OUT (C), r
	case 0x41, 0x49, 0x51, 0x59, 0x61, 0x69, 0x71, 0x79:
		port := uint16(c.B)<<8 | uint16(c.C)
		r := (op >> 3) & 7
		v := byte(0)
		if r != 6 { // 0x71 outputs 0
			v = c.readReg(r)
		}
		c.Out(port, v)
		return 12

	// SBC HL, rr
	case 0x42:
		c.sbcHL16(c.BC())
		return 15
	case 0x52:
		c.sbcHL16(c.DE())
		return 15
	case 0x62:
		c.sbcHL16(c.HL())
		return 15
	case 0x72:
		c.sbcHL16(c.SP)
		return 15

	// ADC HL, rr
	case 0x4A:
		c.adcHL16(c.BC())
		return 15
	case 0x5A:
		c.adcHL16(c.DE())
		return 15
	case 0x6A:
		c.adcHL16(c.HL())
		return 15
	case 0x7A:
		c.adcHL16(c.SP)
		return 15

	// LD (nn), rr
	case 0x43:
		addr := c.fetch16()
		c.write16(addr, c.BC())
		c.WZ = addr + 1
		return 20
	case 0x53:
		addr := c.fetch16()
		c.write16(addr, c.DE())
		c.WZ = addr + 1
		return 20
	case 0x63:
		addr := c.fetch16()
		c.write16(addr, c.HL())
		c.WZ = addr + 1
		return 20
	case 0x73:
		addr := c.fetch16()
		c.write16(addr, c.SP)
		c.WZ = addr + 1
		return 20

	// LD rr, (nn)
	case 0x4B:
		addr := c.fetch16()
		c.SetBC(c.read16(addr))
		c.WZ = addr + 1
		return 20
	case 0x5B:
		addr := c.fetch16()
		c.SetDE(c.read16(addr))
		c.WZ = addr + 1
		return 20
	case 0x6B:
		addr := c.fetch16()
		c.SetHL(c.read16(addr))
		c.WZ = addr + 1
		return 20
	case 0x7B:
		addr := c.fetch16()
		c.SP = c.read16(addr)
		c.WZ = addr + 1
		return 20

	// NEG
	case 0x44, 0x4C, 0x54, 0x5C, 0x64, 0x6C, 0x74, 0x7C:
		a := c.A
		c.A = 0
		c.sub8(a)
		return 8

	// RETN
	case 0x45, 0x55, 0x5D, 0x65, 0x6D, 0x75, 0x7D:
		c.IFF1 = c.IFF2
		c.PC = c.pop()
		return 14

	// RETI
	case 0x4D:
		c.IFF1 = c.IFF2
		c.PC = c.pop()
		return 14

	// IM 0 / IM 1 / IM 2
	case 0x46, 0x66:
		c.IM = 0
		return 8
	case 0x56, 0x76:
		c.IM = 1
		return 8
	case 0x5E, 0x7E:
		c.IM = 2
		return 8

	// LD I, A / LD R, A
	case 0x47:
		c.I = c.A
		return 9
	case 0x4F:
		c.R = c.A
		return 9

	// LD A, I / LD A, R
	case 0x57:
		c.A = c.I
		c.setFlag(FlagS, c.A&0x80 != 0)
		c.setFlag(FlagZ, c.A == 0)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagPV, c.IFF2)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 9
	case 0x5F:
		c.A = c.R
		c.setFlag(FlagS, c.A&0x80 != 0)
		c.setFlag(FlagZ, c.A == 0)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagPV, c.IFF2)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 9

	// RRD / RLD
	case 0x67: // RRD
		hl := c.Mem.Read(c.HL())
		c.Mem.Write(c.HL(), (hl>>4)|(c.A<<4))
		c.A = (c.A & 0xF0) | (hl & 0x0F)
		c.setFlagsSZP(c.A)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		return 18
	case 0x6F: // RLD
		hl := c.Mem.Read(c.HL())
		c.Mem.Write(c.HL(), (hl<<4)|(c.A&0x0F))
		c.A = (c.A & 0xF0) | (hl >> 4)
		c.setFlagsSZP(c.A)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		return 18

	// Block instructions
	case 0xA0: // LDI
		c.ldi()
		return 16
	case 0xA1: // CPI
		c.cpi()
		return 16
	case 0xA2: // INI
		c.ini()
		return 16
	case 0xA3: // OUTI
		c.outi()
		return 16
	case 0xA8: // LDD
		c.ldd()
		return 16
	case 0xA9: // CPD
		c.cpd()
		return 16
	case 0xAA: // IND
		c.ind()
		return 16
	case 0xAB: // OUTD
		c.outd()
		return 16
	case 0xB0: // LDIR
		c.ldi()
		if c.BC() != 0 {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xB1: // CPIR
		c.cpi()
		if c.BC() != 0 && !c.flagZ() {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xB2: // INIR
		c.ini()
		if c.B != 0 {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xB3: // OTIR
		c.outi()
		if c.B != 0 {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xB8: // LDDR
		c.ldd()
		if c.BC() != 0 {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xB9: // CPDR
		c.cpd()
		if c.BC() != 0 && !c.flagZ() {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xBA: // INDR
		c.ind()
		if c.B != 0 {
			c.PC -= 2
			return 21
		}
		return 16
	case 0xBB: // OTDR
		c.outd()
		if c.B != 0 {
			c.PC -= 2
			return 21
		}
		return 16

	default:
		return 8 // NOPI
	}
}

// flagsFromByte computes S,Z,H,P/V,N flags from a byte (for IN instructions).
func flagsFromByte(v byte) byte {
	var f byte
	if v&0x80 != 0 {
		f |= FlagS
	}
	if v == 0 {
		f |= FlagZ
	}
	if parity(v) {
		f |= FlagPV
	}
	f |= v & (FlagF5 | FlagF3)
	return f
}

// Block instruction helpers

func (c *CPU) ldi() {
	v := c.Mem.Read(c.HL())
	c.Mem.Write(c.DE(), v)
	c.SetHL(c.HL() + 1)
	c.SetDE(c.DE() + 1)
	c.SetBC(c.BC() - 1)
	n := c.A + v
	c.setFlag(FlagPV, c.BC() != 0)
	c.setFlag(FlagH, false)
	c.setFlag(FlagN, false)
	c.setFlag(FlagF3, n&0x08 != 0)
	c.setFlag(FlagF5, n&0x02 != 0)
}

func (c *CPU) ldd() {
	v := c.Mem.Read(c.HL())
	c.Mem.Write(c.DE(), v)
	c.SetHL(c.HL() - 1)
	c.SetDE(c.DE() - 1)
	c.SetBC(c.BC() - 1)
	n := c.A + v
	c.setFlag(FlagPV, c.BC() != 0)
	c.setFlag(FlagH, false)
	c.setFlag(FlagN, false)
	c.setFlag(FlagF3, n&0x08 != 0)
	c.setFlag(FlagF5, n&0x02 != 0)
}

func (c *CPU) cpi() {
	v := c.Mem.Read(c.HL())
	result := c.A - v
	c.SetHL(c.HL() + 1)
	c.SetBC(c.BC() - 1)
	c.setFlag(FlagS, result&0x80 != 0)
	c.setFlag(FlagZ, result == 0)
	c.setFlag(FlagH, (c.A&0x0F) < (v&0x0F))
	c.setFlag(FlagPV, c.BC() != 0)
	c.setFlag(FlagN, true)
	n := result
	if c.flagH() {
		n--
	}
	c.setFlag(FlagF3, n&0x08 != 0)
	c.setFlag(FlagF5, n&0x02 != 0)
}

func (c *CPU) cpd() {
	v := c.Mem.Read(c.HL())
	result := c.A - v
	c.SetHL(c.HL() - 1)
	c.SetBC(c.BC() - 1)
	c.setFlag(FlagS, result&0x80 != 0)
	c.setFlag(FlagZ, result == 0)
	c.setFlag(FlagH, (c.A&0x0F) < (v&0x0F))
	c.setFlag(FlagPV, c.BC() != 0)
	c.setFlag(FlagN, true)
	n := result
	if c.flagH() {
		n--
	}
	c.setFlag(FlagF3, n&0x08 != 0)
	c.setFlag(FlagF5, n&0x02 != 0)
}

func (c *CPU) ini() {
	port := uint16(c.B)<<8 | uint16(c.C)
	v := c.In(port)
	c.Mem.Write(c.HL(), v)
	c.SetHL(c.HL() + 1)
	c.B--
	c.setFlag(FlagZ, c.B == 0)
	c.setFlag(FlagN, true)
}

func (c *CPU) ind() {
	port := uint16(c.B)<<8 | uint16(c.C)
	v := c.In(port)
	c.Mem.Write(c.HL(), v)
	c.SetHL(c.HL() - 1)
	c.B--
	c.setFlag(FlagZ, c.B == 0)
	c.setFlag(FlagN, true)
}

func (c *CPU) outi() {
	v := c.Mem.Read(c.HL())
	c.B--
	port := uint16(c.B)<<8 | uint16(c.C)
	c.Out(port, v)
	c.SetHL(c.HL() + 1)
	c.setFlag(FlagZ, c.B == 0)
	c.setFlag(FlagN, true)
}

func (c *CPU) outd() {
	v := c.Mem.Read(c.HL())
	c.B--
	port := uint16(c.B)<<8 | uint16(c.C)
	c.Out(port, v)
	c.SetHL(c.HL() - 1)
	c.setFlag(FlagZ, c.B == 0)
	c.setFlag(FlagN, true)
}

package z80

// execOp executes a single opcode and returns T-states consumed.
func (c *CPU) execOp(op byte) int {
	switch op {
	// NOP
	case 0x00:
		return 4

	// LD rr, nn
	case 0x01: // LD BC, nn
		c.SetBC(c.fetch16())
		return 10
	case 0x11: // LD DE, nn
		c.SetDE(c.fetch16())
		return 10
	case 0x21: // LD HL, nn
		c.SetHL(c.fetch16())
		return 10
	case 0x31: // LD SP, nn
		c.SP = c.fetch16()
		return 10

	// LD (rr), A
	case 0x02: // LD (BC), A
		bc := c.BC()
		c.Mem.Write(bc, c.A)
		c.WZ = uint16(c.A)<<8 | uint16(byte(bc+1))
		return 7
	case 0x12: // LD (DE), A
		de := c.DE()
		c.Mem.Write(de, c.A)
		c.WZ = uint16(c.A)<<8 | uint16(byte(de+1))
		return 7

	// LD A, (rr)
	case 0x0A: // LD A, (BC)
		bc := c.BC()
		c.A = c.Mem.Read(bc)
		c.WZ = bc + 1
		return 7
	case 0x1A: // LD A, (DE)
		de := c.DE()
		c.A = c.Mem.Read(de)
		c.WZ = de + 1
		return 7

	// INC rr
	case 0x03:
		c.SetBC(c.BC() + 1)
		return 6
	case 0x13:
		c.SetDE(c.DE() + 1)
		return 6
	case 0x23:
		c.SetHL(c.HL() + 1)
		return 6
	case 0x33:
		c.SP++
		return 6

	// DEC rr
	case 0x0B:
		c.SetBC(c.BC() - 1)
		return 6
	case 0x1B:
		c.SetDE(c.DE() - 1)
		return 6
	case 0x2B:
		c.SetHL(c.HL() - 1)
		return 6
	case 0x3B:
		c.SP--
		return 6

	// INC r
	case 0x04:
		c.B = c.inc8(c.B)
		return 4
	case 0x0C:
		c.C = c.inc8(c.C)
		return 4
	case 0x14:
		c.D = c.inc8(c.D)
		return 4
	case 0x1C:
		c.E = c.inc8(c.E)
		return 4
	case 0x24:
		c.H = c.inc8(c.H)
		return 4
	case 0x2C:
		c.L = c.inc8(c.L)
		return 4
	case 0x34: // INC (HL)
		addr := c.HL()
		c.Mem.Write(addr, c.inc8(c.Mem.Read(addr)))
		return 11
	case 0x3C:
		c.A = c.inc8(c.A)
		return 4

	// DEC r
	case 0x05:
		c.B = c.dec8(c.B)
		return 4
	case 0x0D:
		c.C = c.dec8(c.C)
		return 4
	case 0x15:
		c.D = c.dec8(c.D)
		return 4
	case 0x1D:
		c.E = c.dec8(c.E)
		return 4
	case 0x25:
		c.H = c.dec8(c.H)
		return 4
	case 0x2D:
		c.L = c.dec8(c.L)
		return 4
	case 0x35: // DEC (HL)
		addr := c.HL()
		c.Mem.Write(addr, c.dec8(c.Mem.Read(addr)))
		return 11
	case 0x3D:
		c.A = c.dec8(c.A)
		return 4

	// LD r, n
	case 0x06:
		c.B = c.fetch()
		return 7
	case 0x0E:
		c.C = c.fetch()
		return 7
	case 0x16:
		c.D = c.fetch()
		return 7
	case 0x1E:
		c.E = c.fetch()
		return 7
	case 0x26:
		c.H = c.fetch()
		return 7
	case 0x2E:
		c.L = c.fetch()
		return 7
	case 0x36: // LD (HL), n
		c.Mem.Write(c.HL(), c.fetch())
		return 10
	case 0x3E:
		c.A = c.fetch()
		return 7

	// Rotate A (simple, no S/Z/P flags)
	case 0x07: // RLCA
		c.A = (c.A << 1) | (c.A >> 7)
		c.setFlag(FlagC, c.A&1 != 0)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 4
	case 0x0F: // RRCA
		carry := c.A & 1
		c.A = (c.A >> 1) | (carry << 7)
		c.setFlag(FlagC, carry != 0)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 4
	case 0x17: // RLA
		carry := byte(0)
		if c.flagC() {
			carry = 1
		}
		c.setFlag(FlagC, c.A&0x80 != 0)
		c.A = (c.A << 1) | carry
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 4
	case 0x1F: // RRA
		carry := byte(0)
		if c.flagC() {
			carry = 0x80
		}
		c.setFlag(FlagC, c.A&1 != 0)
		c.A = (c.A >> 1) | carry
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 4

	// DAA / CPL / SCF / CCF
	case 0x27:
		c.daa()
		return 4
	case 0x2F: // CPL
		c.A ^= 0xFF
		c.setFlag(FlagH, true)
		c.setFlag(FlagN, true)
		c.setFlag(FlagF5, c.A&0x20 != 0)
		c.setFlag(FlagF3, c.A&0x08 != 0)
		return 4
	case 0x37: // SCF
		priorF := c.F
		c.setFlag(FlagC, true)
		c.setFlag(FlagH, false)
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, (priorF|c.A)&0x20 != 0)
		c.setFlag(FlagF3, (priorF|c.A)&0x08 != 0)
		return 4
	case 0x3F: // CCF
		priorF := c.F
		c.setFlag(FlagH, c.flagC())
		c.F ^= FlagC
		c.setFlag(FlagN, false)
		c.setFlag(FlagF5, (priorF|c.A)&0x20 != 0)
		c.setFlag(FlagF3, (priorF|c.A)&0x08 != 0)
		return 4

	// ADD HL, rr
	case 0x09:
		c.WZ = c.HL() + 1
		c.addHL16(c.BC())
		return 11
	case 0x19:
		c.WZ = c.HL() + 1
		c.addHL16(c.DE())
		return 11
	case 0x29:
		c.WZ = c.HL() + 1
		c.addHL16(c.HL())
		return 11
	case 0x39:
		c.WZ = c.HL() + 1
		c.addHL16(c.SP)
		return 11

	// LD (nn), HL and LD HL, (nn)
	case 0x22: // LD (nn), HL
		addr := c.fetch16()
		c.write16(addr, c.HL())
		c.WZ = addr + 1
		return 16
	case 0x2A: // LD HL, (nn)
		addr := c.fetch16()
		c.SetHL(c.read16(addr))
		c.WZ = addr + 1
		return 16

	// LD (nn), A and LD A, (nn)
	case 0x32: // LD (nn), A
		addr := c.fetch16()
		c.Mem.Write(addr, c.A)
		c.WZ = uint16(c.A)<<8 | uint16(byte(addr+1))
		return 13
	case 0x3A: // LD A, (nn)
		addr := c.fetch16()
		c.A = c.Mem.Read(addr)
		c.WZ = addr + 1
		return 13

	// EX AF, AF'
	case 0x08:
		c.A, c.A_ = c.A_, c.A
		c.F, c.F_ = c.F_, c.F
		return 4

	// DJNZ
	case 0x10:
		off := int8(c.fetch())
		c.B--
		if c.B != 0 {
			c.PC = uint16(int(c.PC) + int(off))
			return 13
		}
		return 8

	// JR e
	case 0x18:
		off := int8(c.fetch())
		c.PC = uint16(int(c.PC) + int(off))
		return 12

	// JR cc, e
	case 0x20: // JR NZ, e
		return c.jr(!c.flagZ())
	case 0x28: // JR Z, e
		return c.jr(c.flagZ())
	case 0x30: // JR NC, e
		return c.jr(!c.flagC())
	case 0x38: // JR C, e
		return c.jr(c.flagC())

	// HALT
	case 0x76:
		c.Halted = true
		return 4

	// LD r, r' (0x40–0x7F excluding 0x76)
	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
		0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F,
		0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57,
		0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F,
		0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
		0x68, 0x69, 0x6A, 0x6B, 0x6C, 0x6D, 0x6E, 0x6F,
		0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x77,
		0x78, 0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x7F:
		dst := (op >> 3) & 7
		src := op & 7
		c.writeReg(dst, c.readReg(src))
		if dst == 6 || src == 6 {
			return 7
		}
		return 4

	// ALU operations on A with register
	case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87: // ADD A, r
		c.add8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F: // ADC A, r
		c.adc8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97: // SUB r
		c.sub8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F: // SBC A, r
		c.sbc8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7: // AND r
		c.and8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF: // XOR r
		c.xor8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7: // OR r
		c.or8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4
	case 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF: // CP r
		c.cp8(c.readReg(op & 7))
		if op&7 == 6 {
			return 7
		}
		return 4

	// ALU with immediate
	case 0xC6: // ADD A, n
		c.add8(c.fetch())
		return 7
	case 0xCE: // ADC A, n
		c.adc8(c.fetch())
		return 7
	case 0xD6: // SUB n
		c.sub8(c.fetch())
		return 7
	case 0xDE: // SBC A, n
		c.sbc8(c.fetch())
		return 7
	case 0xE6: // AND n
		c.and8(c.fetch())
		return 7
	case 0xEE: // XOR n
		c.xor8(c.fetch())
		return 7
	case 0xF6: // OR n
		c.or8(c.fetch())
		return 7
	case 0xFE: // CP n
		c.cp8(c.fetch())
		return 7

	// RET cc
	case 0xC0:
		return c.retCC(!c.flagZ())
	case 0xC8:
		return c.retCC(c.flagZ())
	case 0xD0:
		return c.retCC(!c.flagC())
	case 0xD8:
		return c.retCC(c.flagC())
	case 0xE0:
		return c.retCC(!c.flagPV())
	case 0xE8:
		return c.retCC(c.flagPV())
	case 0xF0:
		return c.retCC(!c.flagS())
	case 0xF8:
		return c.retCC(c.flagS())

	// RET
	case 0xC9:
		c.PC = c.pop()
		return 10

	// JP cc, nn
	case 0xC2:
		return c.jpCC(!c.flagZ())
	case 0xCA:
		return c.jpCC(c.flagZ())
	case 0xD2:
		return c.jpCC(!c.flagC())
	case 0xDA:
		return c.jpCC(c.flagC())
	case 0xE2:
		return c.jpCC(!c.flagPV())
	case 0xEA:
		return c.jpCC(c.flagPV())
	case 0xF2:
		return c.jpCC(!c.flagS())
	case 0xFA:
		return c.jpCC(c.flagS())

	// JP nn
	case 0xC3:
		c.PC = c.fetch16()
		c.WZ = c.PC
		return 10

	// JP (HL)
	case 0xE9:
		c.PC = c.HL()
		return 4

	// CALL cc, nn
	case 0xC4:
		return c.callCC(!c.flagZ())
	case 0xCC:
		return c.callCC(c.flagZ())
	case 0xD4:
		return c.callCC(!c.flagC())
	case 0xDC:
		return c.callCC(c.flagC())
	case 0xE4:
		return c.callCC(!c.flagPV())
	case 0xEC:
		return c.callCC(c.flagPV())
	case 0xF4:
		return c.callCC(!c.flagS())
	case 0xFC:
		return c.callCC(c.flagS())

	// CALL nn
	case 0xCD:
		nn := c.fetch16()
		c.push(c.PC)
		c.PC = nn
		c.WZ = nn
		return 17

	// RST p
	case 0xC7:
		c.push(c.PC)
		c.PC = 0x00
		return 11
	case 0xCF:
		c.push(c.PC)
		c.PC = 0x08
		return 11
	case 0xD7:
		c.push(c.PC)
		c.PC = 0x10
		return 11
	case 0xDF:
		c.push(c.PC)
		c.PC = 0x18
		return 11
	case 0xE7:
		c.push(c.PC)
		c.PC = 0x20
		return 11
	case 0xEF:
		c.push(c.PC)
		c.PC = 0x28
		return 11
	case 0xF7:
		c.push(c.PC)
		c.PC = 0x30
		return 11
	case 0xFF:
		c.push(c.PC)
		c.PC = 0x38
		return 11

	// PUSH / POP
	case 0xC1:
		c.SetBC(c.pop())
		return 10
	case 0xD1:
		c.SetDE(c.pop())
		return 10
	case 0xE1:
		c.SetHL(c.pop())
		return 10
	case 0xF1:
		c.SetAF(c.pop())
		return 10
	case 0xC5:
		c.push(c.BC())
		return 11
	case 0xD5:
		c.push(c.DE())
		return 11
	case 0xE5:
		c.push(c.HL())
		return 11
	case 0xF5:
		c.push(c.AF())
		return 11

	// EX DE, HL
	case 0xEB:
		d, e := c.D, c.E
		c.D, c.E = c.H, c.L
		c.H, c.L = d, e
		return 4

	// EX (SP), HL
	case 0xE3:
		hl := c.HL()
		c.SetHL(c.read16(c.SP))
		c.write16(c.SP, hl)
		c.WZ = c.HL()
		return 19

	// EXX
	case 0xD9:
		c.B, c.B_ = c.B_, c.B
		c.C, c.C_ = c.C_, c.C
		c.D, c.D_ = c.D_, c.D
		c.E, c.E_ = c.E_, c.E
		c.H, c.H_ = c.H_, c.H
		c.L, c.L_ = c.L_, c.L
		return 4

	// LD SP, HL
	case 0xF9:
		c.SP = c.HL()
		return 6

	// IN / OUT (simple forms)
	case 0xDB: // IN A, (n)
		port := uint16(c.fetch()) | uint16(c.A)<<8
		c.A = c.In(port)
		return 11
	case 0xD3: // OUT (n), A
		port := uint16(c.fetch()) | uint16(c.A)<<8
		c.Out(port, c.A)
		return 11

	// DI / EI
	case 0xF3: // DI
		c.IFF1 = false
		c.IFF2 = false
		return 4
	case 0xFB: // EI
		c.pendingEI = true
		return 4

	// Prefixes
	case 0xCB:
		return c.execCB()
	case 0xDD:
		return c.execDD()
	case 0xED:
		return c.execED()
	case 0xFD:
		return c.execFD()

	default:
		// Treat undefined opcodes as NOP
		return 4
	}
}

// jr performs a conditional relative jump.
func (c *CPU) jr(cond bool) int {
	off := int8(c.fetch())
	if cond {
		c.PC = uint16(int(c.PC) + int(off))
		return 12
	}
	return 7
}

// jpCC performs a conditional absolute jump.
func (c *CPU) jpCC(cond bool) int {
	nn := c.fetch16()
	if cond {
		c.PC = nn
	}
	return 10
}

// callCC performs a conditional call.
func (c *CPU) callCC(cond bool) int {
	nn := c.fetch16()
	if cond {
		c.push(c.PC)
		c.PC = nn
		return 17
	}
	return 10
}

// retCC performs a conditional return.
func (c *CPU) retCC(cond bool) int {
	if cond {
		c.PC = c.pop()
		return 11
	}
	return 5
}

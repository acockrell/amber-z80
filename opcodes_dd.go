package z80

// execDD handles the 0xDD prefix (IX register instructions).
func (c *CPU) execDD() int {
	return c.execXY(&c.IX)
}

// execXY executes DD/FD-prefixed instructions with the given index register.
// xy points to either IX or IY.
func (c *CPU) execXY(xy *uint16) int {
	op := c.fetch()

	// Helper to get the displacement and compute effective address
	d := func() uint16 {
		off := int8(c.fetch())
		return uint16(int(*xy) + int(off))
	}

	switch op {
	// LD xy, nn
	case 0x21:
		*xy = c.fetch16()
		return 14

	// LD (nn), xy
	case 0x22:
		c.write16(c.fetch16(), *xy)
		return 20

	// LD xy, (nn)
	case 0x2A:
		*xy = c.read16(c.fetch16())
		return 20

	// INC xy / DEC xy
	case 0x23:
		*xy++
		return 10
	case 0x2B:
		*xy--
		return 10

	// ADD xy, rr
	case 0x09:
		c.addXY(xy, c.BC())
		return 15
	case 0x19:
		c.addXY(xy, c.DE())
		return 15
	case 0x29:
		c.addXY(xy, *xy)
		return 15
	case 0x39:
		c.addXY(xy, c.SP)
		return 15

	// LD (xy+d), n
	case 0x36:
		addr := d()
		n := c.fetch()
		c.Mem.Write(addr, n)
		return 19

	// LD r, (xy+d)
	case 0x46:
		c.B = c.Mem.Read(d())
		return 19
	case 0x4E:
		c.C = c.Mem.Read(d())
		return 19
	case 0x56:
		c.D = c.Mem.Read(d())
		return 19
	case 0x5E:
		c.E = c.Mem.Read(d())
		return 19
	case 0x66:
		c.H = c.Mem.Read(d())
		return 19
	case 0x6E:
		c.L = c.Mem.Read(d())
		return 19
	case 0x7E:
		c.A = c.Mem.Read(d())
		return 19

	// LD (xy+d), r
	case 0x70:
		c.Mem.Write(d(), c.B)
		return 19
	case 0x71:
		c.Mem.Write(d(), c.C)
		return 19
	case 0x72:
		c.Mem.Write(d(), c.D)
		return 19
	case 0x73:
		c.Mem.Write(d(), c.E)
		return 19
	case 0x74:
		c.Mem.Write(d(), c.H)
		return 19
	case 0x75:
		c.Mem.Write(d(), c.L)
		return 19
	case 0x77:
		c.Mem.Write(d(), c.A)
		return 19

	// ALU (xy+d)
	case 0x86:
		c.add8(c.Mem.Read(d()))
		return 19
	case 0x8E:
		c.adc8(c.Mem.Read(d()))
		return 19
	case 0x96:
		c.sub8(c.Mem.Read(d()))
		return 19
	case 0x9E:
		c.sbc8(c.Mem.Read(d()))
		return 19
	case 0xA6:
		c.and8(c.Mem.Read(d()))
		return 19
	case 0xAE:
		c.xor8(c.Mem.Read(d()))
		return 19
	case 0xB6:
		c.or8(c.Mem.Read(d()))
		return 19
	case 0xBE:
		c.cp8(c.Mem.Read(d()))
		return 19

	// INC/DEC (xy+d)
	case 0x34:
		addr := d()
		c.Mem.Write(addr, c.inc8(c.Mem.Read(addr)))
		return 23
	case 0x35:
		addr := d()
		c.Mem.Write(addr, c.dec8(c.Mem.Read(addr)))
		return 23

	// LD xyH/xyL, n  (undocumented)
	case 0x26:
		*xy = (*xy & 0x00FF) | uint16(c.fetch())<<8
		return 11
	case 0x2E:
		*xy = (*xy & 0xFF00) | uint16(c.fetch())
		return 11

	// INC/DEC xyH/xyL (undocumented)
	case 0x24:
		hi := c.inc8(byte(*xy >> 8))
		*xy = (*xy & 0x00FF) | uint16(hi)<<8
		return 8
	case 0x25:
		hi := c.dec8(byte(*xy >> 8))
		*xy = (*xy & 0x00FF) | uint16(hi)<<8
		return 8
	case 0x2C:
		lo := c.inc8(byte(*xy))
		*xy = (*xy & 0xFF00) | uint16(lo)
		return 8
	case 0x2D:
		lo := c.dec8(byte(*xy))
		*xy = (*xy & 0xFF00) | uint16(lo)
		return 8

	// LD r, xyH/xyL (undocumented) and LD xyH/xyL, r
	case 0x44:
		c.B = byte(*xy >> 8)
		return 8
	case 0x45:
		c.B = byte(*xy)
		return 8
	case 0x4C:
		c.C = byte(*xy >> 8)
		return 8
	case 0x4D:
		c.C = byte(*xy)
		return 8
	case 0x54:
		c.D = byte(*xy >> 8)
		return 8
	case 0x55:
		c.D = byte(*xy)
		return 8
	case 0x5C:
		c.E = byte(*xy >> 8)
		return 8
	case 0x5D:
		c.E = byte(*xy)
		return 8
	case 0x60:
		*xy = (*xy & 0x00FF) | uint16(c.B)<<8
		return 8
	case 0x61:
		*xy = (*xy & 0x00FF) | uint16(c.C)<<8
		return 8
	case 0x62:
		*xy = (*xy & 0x00FF) | uint16(c.D)<<8
		return 8
	case 0x63:
		*xy = (*xy & 0x00FF) | uint16(c.E)<<8
		return 8
	case 0x64:
		*xy = (*xy & 0x00FF) | (*xy & 0xFF00)
		return 8 // LD xyH, xyH (nop)
	case 0x65:
		*xy = (*xy & 0x00FF) | ((*xy & 0x00FF) << 8)
		return 8 // LD xyH, xyL
	case 0x67:
		*xy = (*xy & 0x00FF) | uint16(c.A)<<8
		return 8
	case 0x68:
		*xy = (*xy & 0xFF00) | uint16(c.B)
		return 8
	case 0x69:
		*xy = (*xy & 0xFF00) | uint16(c.C)
		return 8
	case 0x6A:
		*xy = (*xy & 0xFF00) | uint16(c.D)
		return 8
	case 0x6B:
		*xy = (*xy & 0xFF00) | uint16(c.E)
		return 8
	case 0x6C:
		*xy = (*xy & 0xFF00) | ((*xy & 0xFF00) >> 8)
		return 8 // LD xyL, xyH
	case 0x6D:
		*xy = (*xy & 0xFF00) | (*xy & 0x00FF)
		return 8 // LD xyL, xyL (nop)
	case 0x6F:
		*xy = (*xy & 0xFF00) | uint16(c.A)
		return 8
	case 0x7C:
		c.A = byte(*xy >> 8)
		return 8
	case 0x7D:
		c.A = byte(*xy)
		return 8

	// ALU xyH/xyL (undocumented)
	case 0x84:
		c.add8(byte(*xy >> 8))
		return 8
	case 0x85:
		c.add8(byte(*xy))
		return 8
	case 0x8C:
		c.adc8(byte(*xy >> 8))
		return 8
	case 0x8D:
		c.adc8(byte(*xy))
		return 8
	case 0x94:
		c.sub8(byte(*xy >> 8))
		return 8
	case 0x95:
		c.sub8(byte(*xy))
		return 8
	case 0x9C:
		c.sbc8(byte(*xy >> 8))
		return 8
	case 0x9D:
		c.sbc8(byte(*xy))
		return 8
	case 0xA4:
		c.and8(byte(*xy >> 8))
		return 8
	case 0xA5:
		c.and8(byte(*xy))
		return 8
	case 0xAC:
		c.xor8(byte(*xy >> 8))
		return 8
	case 0xAD:
		c.xor8(byte(*xy))
		return 8
	case 0xB4:
		c.or8(byte(*xy >> 8))
		return 8
	case 0xB5:
		c.or8(byte(*xy))
		return 8
	case 0xBC:
		c.cp8(byte(*xy >> 8))
		return 8
	case 0xBD:
		c.cp8(byte(*xy))
		return 8

	// PUSH / POP xy
	case 0xE5:
		c.push(*xy)
		return 15
	case 0xE1:
		*xy = c.pop()
		return 14

	// EX (SP), xy
	case 0xE3:
		v := c.read16(c.SP)
		c.write16(c.SP, *xy)
		*xy = v
		return 23

	// JP (xy)
	case 0xE9:
		c.PC = *xy
		return 8

	// LD SP, xy
	case 0xF9:
		c.SP = *xy
		return 10

	// DDCB / FDCB prefix
	case 0xCB:
		return c.execXYCB(xy)

	default:
		// Fall back to base opcode (prefix ignored)
		return c.execOp(op)
	}
}

// addXY performs xy = xy + val with carry/half-carry flags (like addHL16 but for IX/IY).
func (c *CPU) addXY(xy *uint16, val uint16) {
	result := uint32(*xy) + uint32(val)
	c.setFlag(FlagC, result > 0xFFFF)
	c.setFlag(FlagH, (*xy&0x0FFF)+(val&0x0FFF) > 0x0FFF)
	c.setFlag(FlagN, false)
	r := uint16(result)
	c.setFlag(FlagF5, r&0x2000 != 0)
	c.setFlag(FlagF3, r&0x0800 != 0)
	*xy = r
}

// execXYCB handles DDCB / FDCB prefixed instructions (indexed bit ops).
func (c *CPU) execXYCB(xy *uint16) int {
	off := int8(c.fetch())
	op := c.fetch()
	addr := uint16(int(*xy) + int(off))
	v := c.Mem.Read(addr)

	n := uint((op >> 3) & 7)
	var result byte

	switch op >> 6 {
	case 0: // rotate/shift (xy+d), result also stored in r if r != 6
		switch op >> 3 {
		case 0:
			result = c.rlc(v)
		case 1:
			result = c.rrc(v)
		case 2:
			result = c.rl(v)
		case 3:
			result = c.rr(v)
		case 4:
			result = c.sla(v)
		case 5:
			result = c.sra(v)
		case 6:
			result = c.sll(v)
		case 7:
			result = c.srl(v)
		}
		c.Mem.Write(addr, result)
		if r := op & 7; r != 6 {
			c.writeReg(r, result)
		}
		return 23
	case 1: // BIT n, (xy+d)
		c.bit(n, v)
		// F3/F5 come from WZ for indexed BIT
		c.setFlag(FlagF5, addr&0x2000 != 0)
		c.setFlag(FlagF3, addr&0x0800 != 0)
		return 20
	case 2: // RES n, (xy+d)
		result = v &^ (1 << n)
		c.Mem.Write(addr, result)
		if r := op & 7; r != 6 {
			c.writeReg(r, result)
		}
		return 23
	case 3: // SET n, (xy+d)
		result = v | (1 << n)
		c.Mem.Write(addr, result)
		if r := op & 7; r != 6 {
			c.writeReg(r, result)
		}
		return 23
	}
	return 23
}

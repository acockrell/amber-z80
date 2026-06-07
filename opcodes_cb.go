package z80

// execCB handles the 0xCB prefix (bit, rotate, shift operations).
func (c *CPU) execCB() int {
	op := c.fetch()
	r := op & 7
	v := c.readReg(r)

	var result byte
	switch op >> 3 {
	case 0: // RLC r
		result = c.rlc(v)
	case 1: // RRC r
		result = c.rrc(v)
	case 2: // RL r
		result = c.rl(v)
	case 3: // RR r
		result = c.rr(v)
	case 4: // SLA r
		result = c.sla(v)
	case 5: // SRA r
		result = c.sra(v)
	case 6: // SLL r (undocumented)
		result = c.sll(v)
	case 7: // SRL r
		result = c.srl(v)
	default: // BIT/RES/SET
		n := uint((op >> 3) & 7)
		switch op >> 6 {
		case 1: // BIT n, r
			if r == 6 {
				c.bitMem(n, v, byte(c.WZ>>8))
				return 12
			}
			c.bit(n, v)
			return 8
		case 2: // RES n, r
			result = v &^ (1 << n)
			c.writeReg(r, result)
			if r == 6 {
				return 15
			}
			return 8
		case 3: // SET n, r
			result = v | (1 << n)
			c.writeReg(r, result)
			if r == 6 {
				return 15
			}
			return 8
		}
		return 8
	}

	c.writeReg(r, result)
	if r == 6 {
		return 15
	}
	return 8
}

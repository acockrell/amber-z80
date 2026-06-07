package z80

// Flag bit positions in the F register.
const (
	FlagC  byte = 1 << 0 // Carry
	FlagN  byte = 1 << 1 // Add/Subtract
	FlagPV byte = 1 << 2 // Parity/Overflow
	FlagF3 byte = 1 << 3 // undocumented (copy of bit 3 of result)
	FlagH  byte = 1 << 4 // Half carry
	FlagF5 byte = 1 << 5 // undocumented (copy of bit 5 of result)
	FlagZ  byte = 1 << 6 // Zero
	FlagS  byte = 1 << 7 // Sign
)

func (c *CPU) flagC() bool  { return c.F&FlagC != 0 }
func (c *CPU) flagN() bool  { return c.F&FlagN != 0 }
func (c *CPU) flagPV() bool { return c.F&FlagPV != 0 }
func (c *CPU) flagH() bool  { return c.F&FlagH != 0 }
func (c *CPU) flagZ() bool  { return c.F&FlagZ != 0 }
func (c *CPU) flagS() bool  { return c.F&FlagS != 0 }

func (c *CPU) setFlag(flag byte, v bool) {
	if v {
		c.F |= flag
	} else {
		c.F &^= flag
	}
}

// parity returns true if the number of set bits in v is even.
func parity(v byte) bool {
	v ^= v >> 4
	v ^= v >> 2
	v ^= v >> 1
	return v&1 == 0
}

// setFlagsSZ sets S and Z flags (and F3/F5 undocumented) from result.
func (c *CPU) setFlagsSZ(result byte) {
	c.setFlag(FlagS, result&0x80 != 0)
	c.setFlag(FlagZ, result == 0)
	c.setFlag(FlagF5, result&0x20 != 0)
	c.setFlag(FlagF3, result&0x08 != 0)
}

// setFlagsSZP sets S, Z, P/V (parity), F3, F5 from result.
func (c *CPU) setFlagsSZP(result byte) {
	c.setFlagsSZ(result)
	c.setFlag(FlagPV, parity(result))
}

// add8 performs A = A + val, sets all flags.
func (c *CPU) add8(val byte) {
	result := uint16(c.A) + uint16(val)
	c.setFlag(FlagC, result > 0xFF)
	c.setFlag(FlagH, (c.A&0x0F)+(val&0x0F) > 0x0F)
	c.setFlag(FlagN, false)
	overflow := ^(c.A ^ val) & (c.A ^ byte(result))
	c.setFlag(FlagPV, overflow&0x80 != 0)
	c.A = byte(result)
	c.setFlagsSZ(c.A)
}

// adc8 performs A = A + val + carry, sets all flags.
func (c *CPU) adc8(val byte) {
	carry := uint16(0)
	if c.flagC() {
		carry = 1
	}
	result := uint16(c.A) + uint16(val) + carry
	hcarry := (c.A & 0x0F) + (val & 0x0F) + byte(carry)
	c.setFlag(FlagH, hcarry > 0x0F)
	c.setFlag(FlagC, result > 0xFF)
	c.setFlag(FlagN, false)
	overflow := ^(c.A ^ val) & (c.A ^ byte(result))
	c.setFlag(FlagPV, overflow&0x80 != 0)
	c.A = byte(result)
	c.setFlagsSZ(c.A)
}

// sub8 performs A = A - val, sets all flags.
func (c *CPU) sub8(val byte) {
	result := uint16(c.A) - uint16(val)
	c.setFlag(FlagC, result > 0xFF)
	c.setFlag(FlagH, (c.A&0x0F) < (val&0x0F))
	c.setFlag(FlagN, true)
	overflow := (c.A ^ val) & (c.A ^ byte(result))
	c.setFlag(FlagPV, overflow&0x80 != 0)
	c.A = byte(result)
	c.setFlagsSZ(c.A)
}

// sbc8 performs A = A - val - carry, sets all flags.
func (c *CPU) sbc8(val byte) {
	carry := uint16(0)
	if c.flagC() {
		carry = 1
	}
	result := uint16(c.A) - uint16(val) - carry
	hcarry := int(c.A&0x0F) - int(val&0x0F) - int(carry)
	c.setFlag(FlagH, hcarry < 0)
	c.setFlag(FlagC, result > 0xFF)
	c.setFlag(FlagN, true)
	overflow := (c.A ^ val) & (c.A ^ byte(result))
	c.setFlag(FlagPV, overflow&0x80 != 0)
	c.A = byte(result)
	c.setFlagsSZ(c.A)
}

// and8 performs A = A & val, sets flags.
func (c *CPU) and8(val byte) {
	c.A &= val
	c.F = FlagH
	c.setFlagsSZP(c.A)
}

// or8 performs A = A | val, sets flags.
func (c *CPU) or8(val byte) {
	c.A |= val
	c.F = 0
	c.setFlagsSZP(c.A)
}

// xor8 performs A = A ^ val, sets flags.
func (c *CPU) xor8(val byte) {
	c.A ^= val
	c.F = 0
	c.setFlagsSZP(c.A)
}

// cp8 compares A with val (like sub8 but doesn't store result).
func (c *CPU) cp8(val byte) {
	saved := c.A
	c.sub8(val)
	c.A = saved
	// F3/F5 come from val in CP, not result
	c.setFlag(FlagF3, val&0x08 != 0)
	c.setFlag(FlagF5, val&0x20 != 0)
}

// inc8 increments v, sets flags (C unchanged).
func (c *CPU) inc8(v byte) byte {
	result := v + 1
	c.setFlag(FlagH, v&0x0F == 0x0F)
	c.setFlag(FlagPV, v == 0x7F)
	c.setFlag(FlagN, false)
	c.setFlagsSZ(result)
	return result
}

// dec8 decrements v, sets flags (C unchanged).
func (c *CPU) dec8(v byte) byte {
	result := v - 1
	c.setFlag(FlagH, v&0x0F == 0x00)
	c.setFlag(FlagPV, v == 0x80)
	c.setFlag(FlagN, true)
	c.setFlagsSZ(result)
	return result
}

// addHL16 performs HL = HL + rr, sets C, H, N flags.
func (c *CPU) addHL16(val uint16) {
	hl := c.HL()
	result := uint32(hl) + uint32(val)
	c.setFlag(FlagC, result > 0xFFFF)
	c.setFlag(FlagH, (hl&0x0FFF)+(val&0x0FFF) > 0x0FFF)
	c.setFlag(FlagN, false)
	r := uint16(result)
	c.setFlag(FlagF5, r&0x2000 != 0)
	c.setFlag(FlagF3, r&0x0800 != 0)
	c.SetHL(r)
}

// adcHL16 performs HL = HL + rr + carry, sets all flags.
func (c *CPU) adcHL16(val uint16) {
	hl := c.HL()
	carry := uint32(0)
	if c.flagC() {
		carry = 1
	}
	result := uint32(hl) + uint32(val) + carry
	c.setFlag(FlagC, result > 0xFFFF)
	c.setFlag(FlagH, (hl&0x0FFF)+(val&0x0FFF)+uint16(carry) > 0x0FFF)
	c.setFlag(FlagN, false)
	r := uint16(result)
	overflow := ^(hl ^ val) & (hl ^ r)
	c.setFlag(FlagPV, overflow&0x8000 != 0)
	c.setFlag(FlagS, r&0x8000 != 0)
	c.setFlag(FlagZ, r == 0)
	c.setFlag(FlagF5, r&0x2000 != 0)
	c.setFlag(FlagF3, r&0x0800 != 0)
	c.SetHL(r)
}

// sbcHL16 performs HL = HL - rr - carry, sets all flags.
func (c *CPU) sbcHL16(val uint16) {
	hl := c.HL()
	carry := uint32(0)
	if c.flagC() {
		carry = 1
	}
	result := uint32(hl) - uint32(val) - carry
	c.setFlag(FlagC, result > 0xFFFF)
	c.setFlag(FlagH, int(hl&0x0FFF)-int(val&0x0FFF)-int(carry) < 0)
	c.setFlag(FlagN, true)
	r := uint16(result)
	overflow := (hl ^ val) & (hl ^ r)
	c.setFlag(FlagPV, overflow&0x8000 != 0)
	c.setFlag(FlagS, r&0x8000 != 0)
	c.setFlag(FlagZ, r == 0)
	c.setFlag(FlagF5, r&0x2000 != 0)
	c.setFlag(FlagF3, r&0x0800 != 0)
	c.SetHL(r)
}

// rlc rotates v left circular, returns result and sets flags.
func (c *CPU) rlc(v byte) byte {
	carry := v >> 7
	result := (v << 1) | carry
	c.F = 0
	c.setFlag(FlagC, carry != 0)
	c.setFlagsSZP(result)
	return result
}

// rrc rotates v right circular.
func (c *CPU) rrc(v byte) byte {
	carry := v & 1
	result := (v >> 1) | (carry << 7)
	c.F = 0
	c.setFlag(FlagC, carry != 0)
	c.setFlagsSZP(result)
	return result
}

// rl rotates v left through carry.
func (c *CPU) rl(v byte) byte {
	oldCarry := byte(0)
	if c.flagC() {
		oldCarry = 1
	}
	result := (v << 1) | oldCarry
	c.F = 0
	c.setFlag(FlagC, v&0x80 != 0)
	c.setFlagsSZP(result)
	return result
}

// rr rotates v right through carry.
func (c *CPU) rr(v byte) byte {
	oldCarry := byte(0)
	if c.flagC() {
		oldCarry = 0x80
	}
	result := (v >> 1) | oldCarry
	c.F = 0
	c.setFlag(FlagC, v&1 != 0)
	c.setFlagsSZP(result)
	return result
}

// sla shifts v left arithmetic (LSB = 0).
func (c *CPU) sla(v byte) byte {
	result := v << 1
	c.F = 0
	c.setFlag(FlagC, v&0x80 != 0)
	c.setFlagsSZP(result)
	return result
}

// sra shifts v right arithmetic (MSB preserved).
func (c *CPU) sra(v byte) byte {
	result := (v >> 1) | (v & 0x80)
	c.F = 0
	c.setFlag(FlagC, v&1 != 0)
	c.setFlagsSZP(result)
	return result
}

// sll shifts v left logical (LSB = 1). Undocumented but widely used.
func (c *CPU) sll(v byte) byte {
	result := (v << 1) | 1
	c.F = 0
	c.setFlag(FlagC, v&0x80 != 0)
	c.setFlagsSZP(result)
	return result
}

// srl shifts v right logical (MSB = 0).
func (c *CPU) srl(v byte) byte {
	result := v >> 1
	c.F = 0
	c.setFlag(FlagC, v&1 != 0)
	c.setFlagsSZP(result)
	return result
}

// bit tests bit n of v, sets Z, H, N flags. F3/F5 come from v.
func (c *CPU) bit(n uint, v byte) {
	c.bitMem(n, v, v)
}

// bitMem tests bit n of v but takes F3/F5 from src (used for BIT n,(HL)
// where src is MEMPTR high byte, and BIT n,(IX+d) where src is high byte
// of the indexed address).
func (c *CPU) bitMem(n uint, v, src byte) {
	c.setFlag(FlagZ, v&(1<<n) == 0)
	c.setFlag(FlagPV, v&(1<<n) == 0)
	c.setFlag(FlagH, true)
	c.setFlag(FlagN, false)
	c.setFlag(FlagS, n == 7 && v&0x80 != 0)
	c.setFlag(FlagF5, src&0x20 != 0)
	c.setFlag(FlagF3, src&0x08 != 0)
}

// daa performs decimal adjust on A.
func (c *CPU) daa() {
	a := c.A
	oldH := c.flagH()
	oldC := c.flagC()
	negative := c.flagN()
	correction := byte(0)
	newC := false
	if oldH || (a&0x0F) > 9 {
		correction |= 0x06
	}
	if oldC || a > 0x99 {
		correction |= 0x60
		newC = true
	}
	if negative {
		c.A = a - correction
	} else {
		c.A = a + correction
	}
	newH := (a^c.A)&0x10 != 0
	c.setFlag(FlagC, newC)
	c.setFlag(FlagH, newH)
	c.setFlag(FlagPV, parity(c.A))
	c.setFlagsSZ(c.A)
}

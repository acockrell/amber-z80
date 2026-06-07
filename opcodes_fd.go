package z80

// execFD handles the 0xFD prefix (IY register instructions).
// Delegates to execXY with the IY register pointer.
func (c *CPU) execFD() int {
	return c.execXY(&c.IY)
}

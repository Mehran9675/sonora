package opusfile

// Ogg's CRC-32: polynomial 0x04c11db7, not reflected, starting at 0.
var crcTable = func() (t [256]uint32) {
	for i := range t {
		r := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if r&0x80000000 != 0 {
				r = r<<1 ^ 0x04c11db7
			} else {
				r <<= 1
			}
		}
		t[i] = r
	}
	return t
}()

func crcUpdate(crc uint32, b []byte) uint32 {
	for _, c := range b {
		crc = crc<<8 ^ crcTable[byte(crc>>24)^c]
	}
	return crc
}

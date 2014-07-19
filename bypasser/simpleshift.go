package bypasser

type shiftCipher struct {
	key byte
}

func NewShiftCipher(key byte) shiftCipher {
	return shiftCipher{key}
}

func (c shiftCipher) Encrypt(dst, src []byte) {
	srcLen := len(src)
	for i := 0; i < srcLen; i++ {
		dst[i] = byte((int(src[i]) + int(c.key)) % 256)
	}
}

func (c shiftCipher) Decrypt(dst, src []byte) {
	srcLen := len(src)
	for i := 0; i < srcLen; i++ {
		dst[i] = byte((int(src[i]) - int(c.key)) % 256)
	}
}

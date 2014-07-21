package bypasser

import (
	"crypto/aes"
	"crypto/cipher"
)

type aESCTRCipher struct {
	aesCipher     cipher.Block
	initialVector []byte
}

func NewAESCTRCipher(key []byte, initialVector []byte) (aESCTRCipher, error) {
	aesBlockEncrypter, err := aes.NewCipher([]byte(key))
	if err != nil {
		return aESCTRCipher{}, err
	} else {
		return aESCTRCipher{aesBlockEncrypter, initialVector}, nil
	}
}

func (c aESCTRCipher) Encrypt(dst, src []byte) {
	aesEncrypter := cipher.NewCTR(c.aesCipher, c.initialVector)
	copy(dst, src)
	aesEncrypter.XORKeyStream(dst, dst)
}

func (c aESCTRCipher) Decrypt(dst, src []byte) {
	aesDecrypter := cipher.NewCTR(c.aesCipher, c.initialVector)
	copy(dst, src)
	aesDecrypter.XORKeyStream(dst, dst)
}

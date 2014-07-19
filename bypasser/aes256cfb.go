package bypasser

import (
	"crypto/aes"
	"crypto/cipher"
)

type aESCFBCipher struct {
	aesCipher     cipher.Block
	initialVector []byte
}

func NewAESCFBCipher(key []byte, initialVector []byte) (aESCFBCipher, error) {
	aesBlockEncrypter, err := aes.NewCipher([]byte(key))
	if err != nil {
		return aESCFBCipher{}, err
	} else {
		return aESCFBCipher{aesBlockEncrypter, initialVector}, nil
	}
}

func (c aESCFBCipher) Encrypt(dst, src []byte) {
	aesEncrypter := cipher.NewCFBEncrypter(c.aesCipher, c.initialVector)
	aesEncrypter.XORKeyStream(dst, src)
}

func (c aESCFBCipher) Decrypt(dst, src []byte) {
	aesDecrypter := cipher.NewCFBDecrypter(c.aesCipher, c.initialVector)
	aesDecrypter.XORKeyStream(dst, src)
}

package bypasser

type GFWCipher interface {
	Encrypt(dst, src []byte)
	Decrypt(dst, src []byte)
}

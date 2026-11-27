//go:build !solution

package blowfish

// #cgo pkg-config: libcrypto
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <openssl/blowfish.h>
import "C"

import "unsafe"

type Blowfish struct {
	key C.BF_KEY
}

func (b *Blowfish) BlockSize() int {
	return int(C.BF_BLOCK)
}

func (b *Blowfish) Encrypt(dst, src []byte) {
	C.BF_ecb_encrypt(
		(*C.uchar)(unsafe.Pointer(&src[0])),
		(*C.uchar)(unsafe.Pointer(&dst[0])),
		&b.key,
		C.BF_ENCRYPT,
	)
}

func (b *Blowfish) Decrypt(dst, src []byte) {
	C.BF_ecb_encrypt(
		(*C.uchar)(unsafe.Pointer(&src[0])),
		(*C.uchar)(unsafe.Pointer(&dst[0])),
		&b.key,
		C.BF_DECRYPT,
	)
}

func New(key []byte) *Blowfish {
	b := &Blowfish{}

	C.BF_set_key(
		&b.key,
		C.int(len(key)),
		(*C.uchar)(unsafe.Pointer(&key[0])),
	)
	return b
}

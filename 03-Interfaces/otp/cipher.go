//go:build !solution

package otp

import (
	"io"
)

type cipherReader struct {
	prng io.Reader
	in   io.Reader
}

type cipherWriter struct {
	prng io.Reader
	out  io.Writer
}

func (c *cipherReader) Read(p []byte) (n int, err error) {
	// read from c.in and put to p
	n, readErr := c.in.Read(p)
	if n > 0 {
		buf := make([]byte, n)
		if _, err = io.ReadFull(c.prng, buf); err != nil {
			return 0, err
		}

		for i := range n {
			p[i] = p[i] ^ buf[i]
		}
	}

	return n, readErr
}

func (c *cipherWriter) Write(p []byte) (n int, err error) {
	// write p to c.out
	buf := make([]byte, len(p))

	if _, err := io.ReadFull(c.prng, buf); err != nil {
		return 0, err
	}

	for i := range buf {
		buf[i] = p[i] ^ buf[i]
	}

	return c.out.Write(buf)
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &cipherReader{
		prng: prng,
		in:   r,
	}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &cipherWriter{
		prng: prng,
		out:  w,
	}
}

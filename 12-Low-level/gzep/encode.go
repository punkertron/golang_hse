//go:build !solution

package gzep

import (
	"compress/gzip"
	"io"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() any {
		ww, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return ww
	},
}

func Encode(data []byte, w io.Writer) error {
	ww := gzipPool.Get().(*gzip.Writer)
	ww.Reset(w)

	defer func() { gzipPool.Put(ww) }()
	defer func() { _ = ww.Close() }()

	if _, err := ww.Write(data); err != nil {
		return err
	}

	return ww.Flush()
}

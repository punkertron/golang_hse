//go:build !solution

package externalsort

import (
	"bufio"
	"io"
	"strings"
)

type lineReader struct {
	reader *bufio.Reader
}

type lineWriter struct {
	out io.Writer
}

func (r *lineReader) ReadLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	return strings.TrimSuffix(line, "\n"), err
}

func (w *lineWriter) Write(l string) error {
	_, err := io.WriteString(w.out, l+"\n")
	return err
}

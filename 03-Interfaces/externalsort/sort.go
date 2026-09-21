//go:build !solution

package externalsort

import (
	"bufio"
	"container/heap"
	"errors"
	"io"
	"os"
	"sort"
)

func NewReader(r io.Reader) LineReader {
	return &lineReader{
		reader: bufio.NewReader(r),
	}
}

func NewWriter(w io.Writer) LineWriter {
	return &lineWriter{
		out: w,
	}
}

func Merge(w LineWriter, readers ...LineReader) error {
	h := &lineHeap{}
	heap.Init(h)

	for _, reader := range readers {
		line, err := reader.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if err == nil || line != "" {
			heap.Push(h, entry{
				line:   line,
				reader: reader,
			})
		}
	}

	for h.Len() > 0 {
		node := heap.Pop(h).(entry)

		if err := w.Write(node.line); err != nil {
			return err
		}

		line, err := node.reader.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if err == nil || line != "" {
			heap.Push(h, entry{
				line:   line,
				reader: node.reader,
			})
		}
	}

	return nil
}

func sortOneFile(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string

	var line string

	reader := NewReader(f)
	for {
		line, err = reader.ReadLine()

		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if err == nil || line != "" {
			lines = append(lines, line)
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}

	sort.Strings(lines)

	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err = f.Truncate(0); err != nil {
		return err
	}

	buffered := bufio.NewWriter(f)
	writer := NewWriter(buffered)

	for _, line := range lines {
		err = writer.Write(line)
		if err != nil {
			return err
		}
	}

	return buffered.Flush()
}

func Sort(w io.Writer, in ...string) error {
	for _, fileName := range in {
		if err := sortOneFile(fileName); err != nil {
			return err
		}
	}

	var readers []LineReader

	for _, fileName := range in {
		f, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer f.Close()

		readers = append(readers, NewReader(f))
	}

	return Merge(NewWriter(w), readers...)
}

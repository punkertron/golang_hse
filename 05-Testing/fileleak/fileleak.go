//go:build !solution

package fileleak

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

type testingT interface {
	Errorf(msg string, args ...interface{})
	Cleanup(func())
}

type fileInfo struct {
	name string
	fd   int
}

const pathToFiles = "/proc/self/fd"

func getOpenFiles() (map[fileInfo]struct{}, error) {
	entries, err := os.ReadDir(pathToFiles)
	if err != nil {
		return nil, err
	}

	files := make(map[fileInfo]struct{}, len(entries))

	for _, entry := range entries {
		fd, err := strconv.Atoi(entry.Name())
		if err != nil {
			return nil, err
		}

		fdPath := filepath.Join(pathToFiles, entry.Name())

		filename, err := os.Readlink(fdPath)
		if errors.Is(err, os.ErrNotExist) {
			continue // Descriptor already closed
		}

		if err != nil {
			return nil, err
		}

		files[fileInfo{filename, fd}] = struct{}{}
	}

	return files, nil
}

func VerifyNone(t testingT) {
	initialFiles, err := getOpenFiles()
	if err != nil {
		t.Errorf("failed to get open filenames: %v", err)
		return
	}

	t.Cleanup(func() {
		currentFiles, err := getOpenFiles()
		if err != nil {
			t.Errorf("Cleanup: failed to get open filenames: %v", err)
			return
		}

		for file := range currentFiles {
			_, ok := initialFiles[file]
			if !ok {
				t.Errorf("leaked file descriptor %d: %s", file.fd, file.name)
			}
		}
	})
}

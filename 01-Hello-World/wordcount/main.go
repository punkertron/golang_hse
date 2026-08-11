//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
)

func processFile(name string, m map[string]int) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		m[line]++
	}

	return scanner.Err()
}

func main() {
	args := os.Args[1:]

	m := make(map[string]int)

	for _, fileName := range args {
		if err := processFile(fileName, m); err != nil {
			panic(err)
		}
	}

	for key, value := range m {
		if value > 1 {
			fmt.Printf("%d\t%s\n", value, key)
		}
	}
}

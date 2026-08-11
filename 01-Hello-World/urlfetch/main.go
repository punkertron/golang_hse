//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func printURLBody(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", body)
}

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		printURLBody(arg)
	}
}

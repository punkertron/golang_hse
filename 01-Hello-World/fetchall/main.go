//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func printURLBody(ch chan struct{}, url string) {
	start := time.Now()

	defer func() { ch <- struct{}{} }()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s: %s\n", url, err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%s: %s\n", url, err.Error())
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("Get %v\t%d\t%s\n", elapsed, len(body), url)
	// fmt.Printf("%s\n", body)
}

func main() {
	start := time.Now()

	args := os.Args[1:]

	channels := make([]chan struct{}, len(args))
	for i, arg := range args {
		channels[i] = make(chan struct{})
		go printURLBody(channels[i], arg)
	}

	for _, ch := range channels {
		<-ch
	}

	elapsed := time.Since(start)
	fmt.Printf("%s elapsed\n", elapsed)
}

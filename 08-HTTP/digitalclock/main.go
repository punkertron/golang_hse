//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	port := flag.Int("port", 0, "Server port (required)")

	flag.Parse()

	if *port < 1 || *port > 65535 {
		log.Fatal("port is required and must be between 1 and 65535")
	}

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, http.HandlerFunc(handler)); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("RemoteAddr =", r.RemoteAddr)

	var err error

	t := time.Now()

	query := r.URL.Query()

	timeStr := query.Get("time")
	if timeStr != "" {
		t, err = time.Parse(time.TimeOnly, timeStr)
		if err != nil || t.Format(time.TimeOnly) != timeStr {
			log.Println(err, timeStr)
			http.Error(w, "Invalid time", http.StatusBadRequest)

			return
		}
	}

	log.Println("time =", t)

	k := 1

	kStr := query.Get("k")
	if kStr != "" {
		k, err = strconv.Atoi(kStr)
		if err != nil {
			log.Println(err)
			http.Error(w, "Invalid k. Should be number or empty", http.StatusBadRequest)

			return
		}

		if k < 1 || k > 30 {
			http.Error(w, "Invalid k. Must be between 1 and 30", http.StatusBadRequest)

			return
		}
	}

	log.Println("k =", k)
	w.Header().Set("Content-Type", "image/png")

	img := drawClock(t.Format(time.TimeOnly), k)
	if err := png.Encode(w, img); err != nil {
		log.Println(err)
	}
}

func drawClock(text string, k int) *image.RGBA {
	symbols := map[rune]string{
		'0': Zero, '1': One, '2': Two, '3': Three,
		'4': Four, '5': Five, '6': Six, '7': Seven,
		'8': Eight, '9': Nine, ':': Colon,
	}

	// one digit: height = 12, width = 8
	// one ':' symbol: height = 12, width = 4
	// totalHeight = 12
	// totalWidth = 6 * 8 + 2 * 4 = 56
	img := image.NewRGBA(image.Rect(0, 0, 56*k, 12*k))

	xOffset := 0

	for _, char := range text {
		rows := strings.Split(symbols[char], "\n")

		for y, row := range rows {
			for x, pixel := range row {
				pixelColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // white
				if pixel == '1' {
					pixelColor = Cyan
				}

				for dy := range k {
					for dx := range k {
						img.Set((xOffset+x)*k+dx, y*k+dy, pixelColor)
					}
				}
			}
		}

		xOffset += len(rows[0])
	}

	return img
}

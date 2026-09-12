//go:build !solution

package reverse

import (
	"strings"
	"unicode/utf8"
)

func Reverse(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	for {
		r, size := utf8.DecodeLastRuneInString(input)
		if size == 0 {
			break
		}

		builder.WriteRune(r)

		input = input[:len(input)-size]
	}

	return builder.String()
}

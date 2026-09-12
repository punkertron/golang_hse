//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
)

func CollapseSpaces(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	needSpace := false

	for _, r := range input {
		if unicode.IsSpace(r) {
			needSpace = true
			continue
		}

		if needSpace {
			builder.WriteByte(' ')

			needSpace = false
		}

		builder.WriteRune(r)
	}

	if needSpace {
		builder.WriteByte(' ')
	}

	return builder.String()
}

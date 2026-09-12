//go:build !solution

package varfmt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func Sprintf(format string, args ...interface{}) string {
	var builder strings.Builder
	builder.Grow(len(format) + len(args)*3)

	cnt := 0

	for i := 0; i < len(format); {
		r, size := utf8.DecodeRuneInString(format[i:])

		if r == '{' {
			add := strings.IndexByte(format[i+1:], '}')
			finish := i + 1 + add

			var num int

			if finish-i == 1 {
				num = cnt
			} else {
				num, _ = strconv.Atoi(format[i+1 : finish])
			}

			i = finish + 1

			fmt.Fprint(&builder, args[num])

			cnt++
		} else {
			builder.WriteRune(r)

			i += size
		}
	}

	return builder.String()
}

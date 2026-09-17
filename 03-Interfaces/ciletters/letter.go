//go:build !solution

package ciletters

import (
	_ "embed"
	"strings"
	"text/template"
)

// lastLines treats a final newline as a line terminator, preserving blank lines.
func lastLines(s string, n int) []string {
	if n <= 0 || s == "" {
		return nil
	}

	s = strings.TrimSuffix(s, "\n")

	end := len(s)
	for i := range n {
		separator := strings.LastIndexByte(s[:end], '\n')
		if separator < 0 {
			return strings.Split(s, "\n")
		}

		if i == n-1 {
			return strings.Split(s[separator+1:], "\n")
		}

		end = separator
	}

	return nil
}

//go:embed letter.tmpl
var letterText string

var letterTemplate = template.Must(template.New("letter").Funcs(template.FuncMap{
	"lastLines": lastLines,
}).Parse(letterText))

func MakeLetter(n *Notification) (string, error) {
	var b strings.Builder
	if err := letterTemplate.Execute(&b, n); err != nil {
		return "", err
	}

	return b.String(), nil
}

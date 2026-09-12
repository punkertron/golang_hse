//go:build !solution

package speller

import "strings"

var numberToWord = map[int64]string{
	0:  "zero",
	1:  "one",
	2:  "two",
	3:  "three",
	4:  "four",
	5:  "five",
	6:  "six",
	7:  "seven",
	8:  "eight",
	9:  "nine",
	10: "ten",
	11: "eleven",
	12: "twelve",
	13: "thirteen",
	14: "fourteen",
	15: "fifteen",
	16: "sixteen",
	17: "seventeen",
	18: "eighteen",
	19: "nineteen",

	20: "twenty",
	30: "thirty",
	40: "forty",
	50: "fifty",
	60: "sixty",
	70: "seventy",
	80: "eighty",
	90: "ninety",
}

type scale struct {
	value int64
	word  string
}

var scales = []scale{
	{value: 1_000_000_000, word: "billion"},
	{value: 1_000_000, word: "million"},
	{value: 1_000, word: "thousand"},
}

func Spell(n int64) string {
	if n == 0 {
		return numberToWord[0]
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var parts []string

	for _, scale := range scales {
		if n < scale.value {
			continue
		}

		parts = append(parts, spellUnderThousand(n/scale.value)+" "+scale.word)
		n %= scale.value
	}

	if n > 0 {
		parts = append(parts, spellUnderThousand(n))
	}

	result := strings.Join(parts, " ")
	if negative {
		return "minus " + result
	}

	return result
}

func spellUnderThousand(n int64) string {
	var parts []string

	if n >= 100 {
		parts = append(parts, numberToWord[n/100]+" hundred")
		n %= 100
	}

	if n >= 20 {
		tens := numberToWord[n/10*10]
		if n%10 != 0 {
			tens += "-" + numberToWord[n%10]
		}

		parts = append(parts, tens)
	} else if n > 0 {
		parts = append(parts, numberToWord[n])
	}

	return strings.Join(parts, " ")
}

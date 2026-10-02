package tabletest

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	parseTests := []struct {
		in      string
		out     time.Duration
		wantErr string
	}{
		{"300ms", 300 * time.Millisecond, ""},
		{"-3ms", -3 * time.Millisecond, ""},
		{"3ns", 3 * time.Nanosecond, ""},
		{"3us", 3 * time.Microsecond, ""},
		{"3µs", 3 * time.Microsecond, ""},
		{"3s", 3 * time.Second, ""},
		{"3m", 3 * time.Minute, ""},
		{"3h", 3 * time.Hour, ""},
		{"", 0, "time: invalid duration "},
		{"-1", 0, "time: missing unit in duration -1"},
		{"3,000,000,000s", 0, "time: unknown unit , in duration 3,000,000,000s"},
		{"30000000000000000000s", 0, "time: invalid duration 30000000000000000000s"},
		{"3.5s", 3500 * time.Millisecond, ""},
		{"0", 0, ""},
		{"s", 0, "time: invalid duration s"},
		{"00s", 0, ""},
		{"-04s", -4 * time.Second, ""},
		{".-s", 0, "time: invalid duration .-s"},
		{"3.50999999999999999999944s", 3510 * time.Millisecond, ""},
		{
			"-30000000000000000000.4s",
			0,
			"time: invalid duration -30000000000000000000.4s",
		},
		{"9223372036854775806.4s", 0, "time: invalid duration 9223372036854775806.4s"},
		{"2h45m", 2*time.Hour + 45*time.Minute, ""},
	}

	for _, test := range parseTests {
		t.Run(test.in, func(t *testing.T) {
			res, err := ParseDuration(test.in)

			sameError := (err == nil && test.wantErr == "") ||
				(err != nil && err.Error() == test.wantErr)

			if res != test.out || !sameError {
				t.Errorf(
					"got %v with err '%v', want %v with err '%v'",
					res,
					err,
					test.out,
					test.wantErr,
				)
			}
		})
	}
}

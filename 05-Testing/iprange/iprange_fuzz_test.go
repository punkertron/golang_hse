package iprange

import (
	"testing"
)

func FuzzParseList(f *testing.F) {
	testcases := []string{
		"10.0.0.1, 10.0.0.5-10, 192.168.1.*, 192.168.10.0/24",
		"55.1574.89",
		"0.0.0.0/70",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, addr string) {
		_, _ = ParseList(addr)
	})
}

func TestParseListInvalidPrefix(t *testing.T) {
	if _, err := ParseList("0.0.0.0/70"); err == nil {
		t.Errorf("expected an error for IPv4 prefix length 70")
	}
}

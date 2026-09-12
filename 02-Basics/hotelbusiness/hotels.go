//go:build !solution

package hotelbusiness

import (
	"sort"
)

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func ComputeLoad(guests []Guest) []Load {
	m := make(map[int]int)

	for _, guest := range guests {
		m[guest.CheckInDate]++
		m[guest.CheckOutDate]--
	}

	dates := make([]int, 0, len(m))
	for date := range m {
		dates = append(dates, date)
	}

	sort.Ints(dates)

	res := make([]Load, 0, len(dates))
	currLoad := 0

	for _, date := range dates {
		change := m[date]
		if change != 0 {
			currLoad += change
			res = append(res, Load{date, currLoad})
		}
	}

	return res
}

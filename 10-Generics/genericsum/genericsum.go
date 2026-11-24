//go:build !solution

package genericsum

import (
	"cmp"
	"sync"
)

func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}

	return b
}

func SortSlice[T cmp.Ordered](a []T) {
	for i := range a {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[j], a[i] = a[i], a[j]
			}
		}
	}
}

func MapsEqual[Key comparable, Value comparable](a, b map[Key]Value) bool {
	type p struct {
		k Key
		v Value
	}

	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if val, ok := b[k]; !ok || val != v {
			return false
		}
	}

	return true
}

func SliceContains[T comparable](s []T, v T) bool {
	for _, val := range s {
		if val == v {
			return true
		}
	}

	return false
}

func MergeChans[T any](chs ...<-chan T) <-chan T {
	out := make(chan T)

	var wg sync.WaitGroup

	for _, ch := range chs {
		wg.Add(1)

		go func(ch <-chan T) {
			defer wg.Done()

			for v := range ch {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func IsHermitianMatrix[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | complex64 | complex128
}](m [][]T) bool {
	for _, row := range m {
		if len(row) != len(m) {
			return false
		}
	}

	for i, row := range m {
		for j := i; j < len(row); j++ {
			transposed := any(m[j][i])
			switch value := transposed.(type) {
			case complex64:
				transposed = complex(real(value), -imag(value))
			case complex128:
				transposed = complex(real(value), -imag(value))
			}

			if any(row[j]) != transposed {
				return false
			}
		}
	}

	return true
}

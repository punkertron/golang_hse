package externalsort

type entry struct {
	line   string
	reader LineReader
}

// lineHeap is a min-heap of entries ordered by line.
type lineHeap []entry

func (h lineHeap) Len() int           { return len(h) }
func (h lineHeap) Less(i, j int) bool { return h[i].line < h[j].line }
func (h lineHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *lineHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(entry))
}

func (h *lineHeap) Pop() any {
	old := *h
	last := len(old) - 1
	x := old[last]
	old[last] = entry{}
	*h = old[:last]

	return x
}

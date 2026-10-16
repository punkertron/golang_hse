//go:build !solution

package tparallel

type T struct {
	parent  *T
	signal  chan struct{}
	done    chan struct{}
	sub     []*T
	barrier chan struct{}
}

func (t *T) Parallel() {
	t.parent.sub = append(t.parent.sub, t)
	t.signal <- struct{}{}

	<-t.parent.barrier // wait until parent's test function returns
}

func (t *T) Run(subtest func(t *T)) {
	subT := &T{
		parent:  t,
		signal:  make(chan struct{}, 1),
		done:    make(chan struct{}),
		barrier: make(chan struct{}),
	}

	go func() {
		subtest(subT)
		close(subT.barrier)

		for _, child := range subT.sub {
			<-child.done
		}

		close(subT.done)
	}()

	select {
	case <-subT.signal:
	case <-subT.done:
	}
}

func Run(topTests []func(t *T)) {
	tParent := T{
		barrier: make(chan struct{}),
	}

	for _, test := range topTests {
		tParent.Run(test)
	}

	close(tParent.barrier)

	for _, sub := range tParent.sub {
		<-sub.done
	}
}

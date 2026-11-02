//go:build !solution

package dupcall

import (
	"context"
	"sync"
)

type Call struct {
	current *execution
	mu      sync.Mutex
}

type execution struct {
	value   any
	err     error
	cancel  context.CancelFunc
	cbDone  chan struct{}
	waiters int
}

func (o *Call) Do(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (result interface{}, err error) {
	e, err := o.join(ctx, cb)
	if err != nil {
		return nil, err
	}
	defer o.release(e)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.cbDone:
		// Closing cbDone publishes the immutable result to every waiter.
		return e.value, e.err
	}
}

func (o *Call) join(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (*execution, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if o.current == nil {
		callbackCtx, cancel := context.WithCancel(context.Background())

		o.current = &execution{
			cbDone: make(chan struct{}),
			cancel: cancel,
		}
		go o.run(callbackCtx, o.current, cb)
	}

	o.current.waiters++

	return o.current, nil
}

func (o *Call) run(
	ctx context.Context,
	e *execution,
	cb func(context.Context) (interface{}, error),
) {
	defer e.cancel()

	value, err := cb(ctx)

	o.mu.Lock()
	defer o.mu.Unlock()

	e.value, e.err = value, err
	if o.current == e {
		o.current = nil
	}

	close(e.cbDone)
}

func (o *Call) release(e *execution) {
	o.mu.Lock()
	defer o.mu.Unlock()

	e.waiters--
	if e.waiters == 0 {
		e.cancel()

		if o.current == e {
			o.current = nil
		}
	}
}

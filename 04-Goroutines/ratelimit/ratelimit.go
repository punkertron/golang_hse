//go:build !solution

package ratelimit

import (
	"context"
	"errors"
	"time"
)

type request struct {
	ctx  context.Context
	resp chan error
}

// Limiter is precise rate limiter with context support.
type Limiter struct {
	req      chan request
	shutdown chan struct{}
}

var ErrStopped = errors.New("limiter stopped")

// NewLimiter returns limiter that throttles rate of successful Acquire() calls
// to maxSize events at any given interval.
func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	l := Limiter{
		req:      make(chan request, 1),
		shutdown: make(chan struct{}),
	}

	go func() {
		var (
			pendingRequests []request
			acquisitions    []time.Time
			timer           *time.Timer
			timerCh         <-chan time.Time
		)

		for {
			cutoff := time.Now().Add(-interval)

			i := 0
			for i < len(acquisitions) && !acquisitions[i].After(cutoff) {
				i++
			}

			acquisitions = acquisitions[i:]

			for len(acquisitions) < maxCount && len(pendingRequests) > 0 {
				req := pendingRequests[0]
				pendingRequests[0] = request{}
				pendingRequests = pendingRequests[1:]

				if req.ctx.Err() == nil {
					acquisitions = append(acquisitions, time.Now())

					req.resp <- nil
				}
			}

			if len(pendingRequests) > 0 {
				delay := time.Until(acquisitions[0].Add(interval))
				if timer == nil {
					timer = time.NewTimer(delay)
				} else {
					timer.Reset(delay)
				}

				timerCh = timer.C
			}

			select {
			case <-l.shutdown:
				if timer != nil {
					timer.Stop()
				}

				return
			case <-timerCh:
				// loop again
			case req := <-l.req:
				if timer != nil {
					timer.Stop()
				}

				pendingRequests = append(pendingRequests, req)
			}
		}
	}()

	return &l
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case <-l.shutdown:
		return ErrStopped
	default:
	}

	req := request{
		ctx:  ctx,
		resp: make(chan error, 1),
	}

	select {
	case <-l.shutdown:
		return ErrStopped
	case <-ctx.Done():
		return ctx.Err()
	case l.req <- req:
		// manager accepts request
	}

	select {
	case <-l.shutdown:
		return ErrStopped
	case <-ctx.Done():
		return ctx.Err()
	case err := <-req.resp:
		return err
	}
}

func (l *Limiter) Stop() {
	close(l.shutdown)
}

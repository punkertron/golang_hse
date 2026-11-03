//go:build !solution

package pubsub

import (
	"context"
	"errors"
	"slices"
	"sync"
)

/*
Basic architecture:
- MyPubSub stores all active subscriptions (MySubscription).
- For every subscription, one worker goroutine is created.
- Each subscription has its own queue of messages. MyPubSub.Publish adds
  a message to the queue. The subscription's worker processes this queue.
- MyPubSub.Close waits for all workers to drain their queues and finish.
  If its context is canceled, Close returns while workers continue processing.
*/

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	queue    []any         // queue of messages
	cb       MsgHandler    // callback that processes each queued message
	subj     string        // key in MyPubSub.subs
	pubsub   *MyPubSub     // pointer to the owning pubsub
	done     chan struct{} // requests stopping without draining; a running callback finishes
	run      chan struct{} // wakes the worker to check its queue
	once     sync.Once     // makes repeated Unsubscribe calls safe
	mu       sync.Mutex    // protects queue and draining
	draining bool          // if true, the worker processes remaining messages and exits
}

func (s *MySubscription) Unsubscribe() {
	s.once.Do(func() {
		close(s.done)

		s.pubsub.mu.Lock()
		defer s.pubsub.mu.Unlock()

		subArr := s.pubsub.subs[s.subj]
		for i, sub := range subArr {
			if s == sub {
				s.pubsub.subs[s.subj] = slices.Delete(subArr, i, i+1)
				break
			}
		}

		s.mu.Lock()
		s.queue = nil
		s.mu.Unlock()
	})
}

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	workers        sync.WaitGroup               // tracks workers until they finish
	closedFinished chan struct{}                // closed when all workers have finished during shutdown
	mu             sync.Mutex                   // protects subs and closed
	subs           map[string][]*MySubscription // map of subscriptions
	closed         bool                         // marks that shutdown has started
}

func NewPubSub() PubSub {
	return &MyPubSub{
		closedFinished: make(chan struct{}),
		subs:           make(map[string][]*MySubscription),
	}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	sub := &MySubscription{
		queue:  make([]any, 0),
		cb:     cb,
		subj:   subj,
		pubsub: p,
		done:   make(chan struct{}),
		run:    make(chan struct{}, 1),
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("subscribe: pubsub is already closed")
	}

	p.workers.Add(1)
	p.subs[subj] = append(p.subs[subj], sub)
	p.mu.Unlock()

	p.runWorker(sub)

	return sub, nil
}

func (p *MyPubSub) runWorker(sub *MySubscription) {
	go func() {
		defer p.workers.Done()

		for {
			select {
			case <-sub.done:
				return
			case <-sub.run:
				sub.mu.Lock()
				for len(sub.queue) > 0 {
					select {
					case <-sub.done:
						sub.mu.Unlock()
						return
					default:
					}

					msg := sub.queue[0]
					sub.queue[0] = nil
					sub.queue = sub.queue[1:]
					sub.mu.Unlock()
					sub.cb(msg)

					sub.mu.Lock()
				}

				if sub.draining {
					sub.mu.Unlock()
					return
				}
				sub.mu.Unlock()
			}
		}
	}()
}

func (s *MySubscription) start() {
	select {
	case s.run <- struct{}{}:
	default:
	}
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return errors.New("publish: pubsub is already closed")
	}

	for _, sub := range p.subs[subj] {
		sub.mu.Lock()
		sub.queue = append(sub.queue, msg)
		sub.mu.Unlock()
		sub.start()
	}

	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
	} else {
		p.closed = true

		for _, subArr := range p.subs {
			for _, sub := range subArr {
				sub.mu.Lock()
				sub.draining = true
				sub.start() // wake up worker
				sub.mu.Unlock()
			}
		}
		p.mu.Unlock()

		go func() {
			p.workers.Wait()
			close(p.closedFinished)
		}()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closedFinished:
		return nil
	}
}

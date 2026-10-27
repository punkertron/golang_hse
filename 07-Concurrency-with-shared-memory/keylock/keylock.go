//go:build !solution

package keylock

type KeyLock struct {
	locker        chan struct{}
	acquiredLocks map[string]chan struct{}
}

func New() *KeyLock {
	return &KeyLock{
		locker:        make(chan struct{}, 1),
		acquiredLocks: make(map[string]chan struct{}),
	}
}

func hasDuplicates(keys []string) bool {
	m := make(map[string]struct{})

	for _, key := range keys {
		if _, exists := m[key]; exists {
			return true
		}

		m[key] = struct{}{}
	}

	return false
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	if hasDuplicates(keys) {
		panic("Duplicate keys")
	}

	acquiredKeys := make([]string, 0, len(keys))

	unlock = func() {
		l.locker <- struct{}{}

		for _, key := range acquiredKeys {
			ch := l.acquiredLocks[key]
			delete(l.acquiredLocks, key)
			close(ch)
		}

		<-l.locker

		acquiredKeys = acquiredKeys[:0]
	}

	for i := 0; i < len(keys); {
		// cancel or lock the map
		select {
		case <-cancel:
			unlock()

			canceled = true

			return
		case l.locker <- struct{}{}:
		}

		// double check cancel, because cancel and l.locker can be both ready
		// at the same time and the order is not guaranteed. Already cancelled
		// single-key request can therefore succeed
		select {
		case <-cancel:
			<-l.locker
			unlock()

			canceled = true

			return
		default:
		}

		key := keys[i]

		if acq, ok := l.acquiredLocks[key]; ok {
			// already acquired
			<-l.locker
			unlock()

			i = 0

			select {
			case <-acq:
			case <-cancel:
				canceled = true
				return
			}
		} else {
			l.acquiredLocks[key] = make(chan struct{})
			acquiredKeys = append(acquiredKeys, key)
			i++

			<-l.locker
		}
	}

	return
}

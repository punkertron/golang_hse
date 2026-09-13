//go:build !solution

package lrucache

import (
	"container/list"
)

type pair struct {
	key, value int
}

type lruCache struct {
	maxCapacity int
	m           map[int]*list.Element
	l           list.List // *pair. recent item is in the back
}

// Get returns value associated with the key.
//
// The second value is a bool that is true if the key exists in the cache,
// and false if not.
func (c *lruCache) Get(key int) (int, bool) {
	elem, ok := c.m[key]
	if !ok {
		return 0, false
	}

	c.l.MoveToBack(elem)

	return elem.Value.(*pair).value, true
}

// Set updates value associated with the key.
//
// If there is no key in the cache new (key, value) pair is created.
func (c *lruCache) Set(key, value int) {
	if c.maxCapacity == 0 {
		return
	}

	elem, ok := c.m[key]
	if ok {
		c.l.MoveToBack(elem)
		elem.Value.(*pair).value = value

		return
	}

	if c.maxCapacity == c.l.Len() {
		oldest := c.l.Front()
		p := oldest.Value.(*pair)
		delete(c.m, p.key)

		p.key, p.value = key, value

		c.l.MoveToBack(oldest)
		c.m[key] = c.l.Back()

		return
	}

	c.m[key] = c.l.PushBack(&pair{key, value})
}

// Range calls function f on all elements of the cache
// in increasing access time order.
//
// Stops earlier if f returns false.
func (c *lruCache) Range(f func(key, value int) bool) {
	for e := c.l.Front(); e != nil; e = e.Next() {
		p := e.Value.(*pair)
		if !f(p.key, p.value) {
			return
		}
	}
}

// Clear removes all keys and values from the cache.
func (c *lruCache) Clear() {
	c.l.Init()
	c.m = make(map[int]*list.Element)
}

func New(maxCapacity int) Cache {
	return &lruCache{
		maxCapacity: maxCapacity,
		m:           make(map[int]*list.Element, maxCapacity),
	}
}

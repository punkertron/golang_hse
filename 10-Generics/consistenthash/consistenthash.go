//go:build !solution

package consistenthash

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strconv"
)

const virtualNodesPerNode = 256

type Node interface {
	// ID is some persistent and unique identifier
	ID() string
}

type virtualNode[N Node] struct {
	position uint64
	node     *N
}

type ConsistentHash[N Node] struct {
	nodes map[string]*N
	ring  []virtualNode[N]
}

func New[N Node]() *ConsistentHash[N] {
	return &ConsistentHash[N]{}
}

func (h *ConsistentHash[N]) AddNode(n *N) {
	if h.nodes == nil {
		h.nodes = make(map[string]*N)
	}

	h.nodes[(*n).ID()] = n
	h.ring = nil
}

func (h *ConsistentHash[N]) RemoveNode(n *N) {
	delete(h.nodes, (*n).ID())
	h.ring = nil
}

func (h *ConsistentHash[N]) GetNode(key string) *N {
	if len(h.nodes) == 0 {
		return nil
	}

	if h.ring == nil {
		h.rebuildRing()
	}

	index, _ := slices.BinarySearchFunc(
		h.ring,
		hashPosition(key),
		func(point virtualNode[N], position uint64) int {
			return cmp.Compare(point.position, position)
		},
	)
	if index == len(h.ring) {
		index = 0
	}

	return h.ring[index].node
}

// Rebuild once on the next lookup after one or more membership changes
func (h *ConsistentHash[N]) rebuildRing() {
	h.ring = make([]virtualNode[N], 0, len(h.nodes)*virtualNodesPerNode)
	for id, node := range h.nodes {
		for replica := range virtualNodesPerNode {
			position := hashPosition(strconv.Itoa(replica) + ":" + id)
			h.ring = append(h.ring, virtualNode[N]{position: position, node: node})
		}
	}

	slices.SortFunc(h.ring, func(a, b virtualNode[N]) int {
		if a.position == b.position {
			return cmp.Compare((*a.node).ID(), (*b.node).ID())
		}

		return cmp.Compare(a.position, b.position)
	})
}

func hashPosition(key string) uint64 {
	sum := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(sum[:8])
}

//go:build !solution

package httpgauge

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"sync"

	"github.com/go-chi/chi/v5"
)

type Gauge struct {
	metrics map[string]int
	mu      sync.Mutex
}

func New() *Gauge {
	return &Gauge{
		metrics: make(map[string]int),
	}
}

func (g *Gauge) Snapshot() map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	copied := maps.Clone(g.metrics)

	return copied
}

// ServeHTTP returns accumulated statistics in text format ordered by pattern.
//
// For example:
//
//	/a 10
//	/b 5
//	/c/{id} 7
func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := g.Snapshot()

	keys := slices.Sorted(maps.Keys(metrics))
	for _, key := range keys {
		fmt.Fprintf(w, "%s %d\n", key, metrics[key])
	}
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ctx := chi.RouteContext(r.Context())
			if ctx != nil {
				path := ctx.RoutePattern()

				g.mu.Lock()
				defer g.mu.Unlock()

				g.metrics[path]++
			}
		}()

		next.ServeHTTP(w, r)
	})
}

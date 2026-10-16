package realtime

import (
	"strconv"
	"sync"

	"maidcafe-engine/app/models"
)

type Hub struct {
	mu     sync.Mutex
	subs   map[string]map[chan models.Message]struct{}
	done   chan struct{}
	closed bool
}

func NewHub() *Hub {
	return &Hub{subs: map[string]map[chan models.Message]struct{}{}, done: make(chan struct{})}
}

func ConvKey(convID int64) string { return strconv.FormatInt(convID, 10) }

func (h *Hub) Subscribe(key string) (<-chan models.Message, func()) {
	ch := make(chan models.Message, 32)
	h.mu.Lock()
	if h.subs[key] == nil {
		h.subs[key] = map[chan models.Message]struct{}{}
	}
	h.subs[key][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs[key], ch)
		if len(h.subs[key]) == 0 {
			delete(h.subs, key)
		}
		h.mu.Unlock()
	}
}

func (h *Hub) Publish(key string, msgs ...models.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, m := range msgs {
		for ch := range h.subs[key] {
			select {
			case ch <- m:
			default:
			}
		}
	}
}

func (h *Hub) Done() <-chan struct{} { return h.done }

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.closed {
		h.closed = true
		close(h.done)
	}
}

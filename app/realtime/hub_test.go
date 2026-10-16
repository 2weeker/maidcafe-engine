package realtime

import (
	"testing"
	"time"

	"maidcafe-engine/app/models"
)

func TestPublishReachesOnlyThatConversation(t *testing.T) {
	h := NewHub()
	a, unsubA := h.Subscribe(ConvKey(12))
	b, unsubB := h.Subscribe(ConvKey(12))
	other, unsubOther := h.Subscribe(ConvKey(13))
	defer unsubOther()

	h.Publish(ConvKey(12), models.Message{ID: 7, Body: "hi"})
	for _, ch := range []<-chan models.Message{a, b} {
		select {
		case m := <-ch:
			if m.ID != 7 {
				t.Fatalf("got %+v", m)
			}
		case <-time.After(time.Second):
			t.Fatal("subscriber didn't receive")
		}
	}
	select {
	case m := <-other:
		t.Fatalf("other conversation received %+v", m)
	default:
	}

	unsubA()
	unsubB()
	h.mu.Lock()
	n := len(h.subs)
	h.mu.Unlock()
	if n != 1 {
		t.Fatalf("want 1 key left, got %d", n)
	}
}

func TestPublishManyKeepsOrder(t *testing.T) {
	h := NewHub()
	ch, unsub := h.Subscribe("k")
	defer unsub()
	h.Publish("k", models.Message{ID: 1}, models.Message{ID: 2})
	if (<-ch).ID != 1 || (<-ch).ID != 2 {
		t.Fatal("order lost")
	}
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	h := NewHub()
	_, unsub := h.Subscribe("k")
	defer unsub()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			h.Publish("k", models.Message{ID: int64(i)})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on a full subscriber")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	h := NewHub()
	h.Close()
	h.Close()
	select {
	case <-h.Done():
	default:
		t.Fatal("done not closed")
	}
}

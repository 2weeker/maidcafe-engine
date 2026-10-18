package limits

import (
	"math"
	"net/netip"
	"sync"
	"time"
)

const forgetAfter = time.Hour

const sweepAbove = 10000

func AddressKey(ip string) string {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ip
	}
	addr = addr.Unmap()
	if addr.Is4() {
		return addr.String()
	}
	prefix, err := addr.Prefix(64)
	if err != nil {
		return ip
	}
	return prefix.String()
}

type Rate struct {
	mu      sync.Mutex
	burst   float64
	perSec  float64
	buckets map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewRate(requests int, per time.Duration) *Rate {
	if requests <= 0 || per <= 0 {
		return nil
	}
	return &Rate{
		burst:   float64(requests),
		perSec:  float64(requests) / per.Seconds(),
		buckets: map[string]*bucket{},
		now:     time.Now,
	}
}

func (r *Rate) Allow(key string) (time.Duration, bool) {
	if r == nil {
		return 0, true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	b, ok := r.buckets[key]
	if !ok {
		r.sweep(now)
		b = &bucket{tokens: r.burst, last: now}
		r.buckets[key] = b
	}
	b.tokens = math.Min(r.burst, b.tokens+now.Sub(b.last).Seconds()*r.perSec)
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return 0, true
	}
	wait := time.Duration((1 - b.tokens) / r.perSec * float64(time.Second))
	return max(wait, time.Second), false
}

func (r *Rate) Ready(key string) (time.Duration, bool) {
	if r == nil {
		return 0, true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.buckets[key]
	if !ok {
		return 0, true
	}
	tokens := math.Min(r.burst, b.tokens+r.now().Sub(b.last).Seconds()*r.perSec)
	if tokens >= 1 {
		return 0, true
	}
	return max(time.Duration((1-tokens)/r.perSec*float64(time.Second)), time.Second), false
}

func (r *Rate) sweep(now time.Time) {
	if len(r.buckets) < sweepAbove {
		return
	}
	full := time.Duration(r.burst / r.perSec * float64(time.Second))
	for k, b := range r.buckets {
		if now.Sub(b.last) >= full {
			delete(r.buckets, k)
		}
	}
}

type Failures struct {
	mu      sync.Mutex
	limit   int
	lockout time.Duration
	counts  map[string]*failures
	now     func() time.Time
}

type failures struct {
	count int
	last  time.Time
	until time.Time
}

func NewFailures(limit int, lockout time.Duration) *Failures {
	if limit < 1 {
		limit = 1
	}
	return &Failures{limit: limit, lockout: lockout, counts: map[string]*failures{}, now: time.Now}
}

func (f *Failures) Locked(addr string) (time.Duration, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	st, ok := f.counts[addr]
	if now := f.now(); ok && now.Before(st.until) {
		return st.until.Sub(now), true
	}
	return 0, false
}

func (f *Failures) Fail(addr string) time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := f.now()
	st := f.counts[addr]
	switch {
	case st != nil && now.Before(st.until):
		return 0
	case st == nil || now.Sub(st.last) > forgetAfter:
		st = &failures{}
		f.counts[addr] = st
	}
	st.count++
	st.last = now

	if len(f.counts) > sweepAbove {
		for k, v := range f.counts {
			if now.Sub(v.last) > forgetAfter && !now.Before(v.until) {
				delete(f.counts, k)
			}
		}
	}
	if st.count < f.limit || f.lockout <= 0 {
		return 0
	}
	st.count = 0
	st.until = now.Add(f.lockout)
	return f.lockout
}

func (f *Failures) Succeed(addr string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if st, ok := f.counts[addr]; ok && !f.now().Before(st.until) {
		delete(f.counts, addr)
	}
}

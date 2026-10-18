package limits

import (
	"testing"
	"time"
)

func clock() (func() time.Time, *time.Time) {
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return now }, &now
}

func TestAddressKey(t *testing.T) {
	for in, want := range map[string]string{
		"203.0.113.7":                  "203.0.113.7",
		"::ffff:203.0.113.7":           "203.0.113.7",
		"2001:db8:1:2:aaaa::1":         "2001:db8:1:2::/64",
		"2001:db8:1:2:bbbb:cccc::9999": "2001:db8:1:2::/64",
		"not an ip":                    "not an ip",
	} {
		if got := AddressKey(in); got != want {
			t.Errorf("AddressKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRateAllowsABurstThenRefills(t *testing.T) {
	r := NewRate(3, 3*time.Second)
	now, at := clock()
	r.now = now
	for i := 0; i < 3; i++ {
		if _, ok := r.Allow("a"); !ok {
			t.Fatalf("request %d of the burst refused", i+1)
		}
	}
	wait, ok := r.Allow("a")
	if ok || wait < time.Second {
		t.Fatalf("the fourth request must wait about a second, got ok=%v wait=%v", ok, wait)
	}
	if _, ok := r.Allow("b"); !ok {
		t.Error("another key has its own bucket")
	}
	*at = at.Add(time.Second)
	if _, ok := r.Allow("a"); !ok {
		t.Error("one token refills after a second")
	}
	if _, ok := r.Allow("a"); ok {
		t.Error("only one token refilled")
	}
	*at = at.Add(time.Hour)
	for i := 0; i < 3; i++ {
		if _, ok := r.Allow("a"); !ok {
			t.Fatal("refill is capped at the burst, and a full burst is back")
		}
	}
	if _, ok := r.Allow("a"); ok {
		t.Error("refill must not exceed the burst")
	}
}

func TestDisabledRateAllowsEverything(t *testing.T) {
	for _, r := range []*Rate{NewRate(0, time.Minute), NewRate(5, 0)} {
		if r != nil {
			t.Fatal("zero requests or zero window disables the limit")
		}
		for i := 0; i < 1000; i++ {
			if _, ok := r.Allow("a"); !ok {
				t.Fatal("a disabled limit refused a request")
			}
		}
	}
}

func TestFailuresLockOut(t *testing.T) {
	f := NewFailures(3, time.Minute)
	now, at := clock()
	f.now = now
	if f.Fail("a") != 0 || f.Fail("a") != 0 {
		t.Fatal("the first two failures don't lock")
	}
	if lock := f.Fail("a"); lock != time.Minute {
		t.Fatalf("the third failure locks for a minute, got %v", lock)
	}
	if wait, locked := f.Locked("a"); !locked || wait != time.Minute {
		t.Fatalf("locked for a minute, got %v %v", wait, locked)
	}
	if _, locked := f.Locked("b"); locked {
		t.Error("another address isn't locked")
	}
	*at = at.Add(time.Minute)
	if _, locked := f.Locked("a"); locked {
		t.Error("the lock expires")
	}
	f.Fail("a")
	f.Succeed("a")
	f.Fail("a")
	if lock := f.Fail("a"); lock != 0 {
		t.Error("a success resets the count")
	}
}

func TestFailuresAreForgotten(t *testing.T) {
	f := NewFailures(2, time.Minute)
	now, at := clock()
	f.now = now
	f.Fail("a")
	*at = at.Add(forgetAfter + time.Second)
	if lock := f.Fail("a"); lock != 0 {
		t.Error("a failure older than an hour no longer counts")
	}
}

func TestReadyDoesNotSpend(t *testing.T) {
	r := NewRate(1, time.Minute)
	now, _ := clock()
	r.now = now
	for i := 0; i < 5; i++ {
		if _, ok := r.Ready("a"); !ok {
			t.Fatal("checking must not use up the token")
		}
	}
	r.Allow("a")
	if wait, ok := r.Ready("a"); ok || wait < time.Second {
		t.Errorf("after spending the only token, Ready reports a wait: %v %v", wait, ok)
	}
}

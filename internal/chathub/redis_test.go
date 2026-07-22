package chathub

import (
	"sync"
	"testing"

	"github.com/google/uuid"
)

// newTestBridge builds a Bridge wired only for the queue mechanics under test.
// The Redis client and hub are unused here (no Run loop, no I/O), so nil-ish
// zero values are fine — these tests exercise enqueue/takePending only.
func newTestBridge() *Bridge {
	return &Bridge{logger: testLogger(), wake: make(chan struct{}, 1)}
}

// TestEnqueue_NeverBlocks is the core deadlock guard: enqueue is called under
// the hub write lock, so it must return without parking even when nothing is
// draining the queue and far more commands arrive than the old fixed 1024-slot
// buffer could hold. A regression to a bounded blocking channel would hang here.
func TestEnqueue_NeverBlocks(t *testing.T) {
	b := newTestBridge()
	const n = 5000 // > old cmdBuffer (1024): a bounded channel would block
	done := make(chan struct{})
	go func() {
		for i := range n {
			b.enqueue(convChannelPrefix+uuid.New().String(), i%2 == 0)
		}
		close(done)
	}()
	<-done // if enqueue blocked, this test times out (deadlock detector fires)

	if got := len(b.takePending()); got != n {
		t.Fatalf("takePending returned %d commands, want %d", got, n)
	}
}

// TestEnqueue_PreservesOrder verifies the handoff neither drops nor reorders:
// takePending returns the exact sequence enqueue appended, which is what keeps
// the subscribe/unsubscribe refcount correct.
func TestEnqueue_PreservesOrder(t *testing.T) {
	b := newTestBridge()
	ch := convChannelPrefix + uuid.New().String()
	want := []bool{true, false, true, true, false}
	for _, sub := range want {
		b.enqueue(ch, sub)
	}
	got := b.takePending()
	if len(got) != len(want) {
		t.Fatalf("got %d commands, want %d", len(got), len(want))
	}
	for i, cmd := range got {
		if cmd.channel != ch || cmd.subscribe != want[i] {
			t.Fatalf("command %d = {%s, %v}, want {%s, %v}", i, cmd.channel, cmd.subscribe, ch, want[i])
		}
	}
}

// TestEnqueue_ConcurrentAppendsAreRaceClean exercises enqueue and takePending
// concurrently: they share the pending slice, so their locking must hold up
// under -race (a drainer racing several producers, as Run races the hub hooks).
func TestEnqueue_ConcurrentAppendsAreRaceClean(t *testing.T) {
	b := newTestBridge()

	var producers sync.WaitGroup
	for range 8 {
		producers.Go(func() {
			for range 500 {
				b.enqueue(userChannelPrefix+uuid.New().String(), true)
			}
		})
	}

	stop := make(chan struct{})
	var drainer sync.WaitGroup
	drainer.Go(func() {
		for {
			select {
			case <-stop:
				b.takePending() // final sweep
				return
			case <-b.wake:
				b.takePending()
			}
		}
	})

	producers.Wait()
	close(stop)
	drainer.Wait()
}

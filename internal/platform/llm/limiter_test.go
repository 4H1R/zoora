package llm_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/llm"
)

type blockingLLM struct {
	inFlight int32
	maxSeen  int32
	release  chan struct{}
}

func (b *blockingLLM) Generate(ctx context.Context, req domain.LLMRequest) (domain.LLMResponse, error) {
	n := atomic.AddInt32(&b.inFlight, 1)
	for {
		old := atomic.LoadInt32(&b.maxSeen)
		if n <= old || atomic.CompareAndSwapInt32(&b.maxSeen, old, n) {
			break
		}
	}
	<-b.release
	atomic.AddInt32(&b.inFlight, -1)
	return domain.LLMResponse{Text: "ok"}, nil
}

func TestLimiterCapsConcurrency(t *testing.T) {
	inner := &blockingLLM{release: make(chan struct{})}
	limited := llm.NewLimiter(inner, 2)

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = limited.Generate(context.Background(), domain.LLMRequest{})
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(inner.release)
	wg.Wait()

	if got := atomic.LoadInt32(&inner.maxSeen); got > 2 {
		t.Fatalf("concurrency exceeded cap: saw %d in flight", got)
	}
}

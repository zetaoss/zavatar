// infra/cloudflare/batch_purger.go
package cloudflare

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	maxPrefixesPerRequest = 30
	flushInterval         = 15 * time.Second
	reqTimeout            = 10 * time.Second
	queueSize             = 1000

	maxAttempts       = 5
	baseRetryDelay    = 500 * time.Millisecond
	maxConcurrentReqs = 5
)

type BatchPurger struct {
	client Purger
	queue  chan string
	sem    chan struct{}
	wg     sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func NewBatchPurger(client Purger) *BatchPurger {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	bp := &BatchPurger{
		client: client,
		queue:  make(chan string, queueSize),
		sem:    make(chan struct{}, maxConcurrentReqs),
		ctx:    ctx,
		cancel: cancel,
	}
	bp.wg.Add(1)
	go bp.run()
	return bp
}

func (b *BatchPurger) Add(prefix string) {
	if prefix == "" {
		return
	}
	select {
	case b.queue <- prefix:
	case <-b.ctx.Done():
		slog.Warn("batch purger closing, dropping prefix", "prefix", prefix)
	default:
		slog.Warn("batch purger queue full, dropping prefix", "prefix", prefix, "queue_size", queueSize)
	}
}

func (b *BatchPurger) Close() {
	slog.Info("batch purger shutting down")
	b.cancel()
	b.wg.Wait()
	slog.Info("batch purger stopped")
}

func (b *BatchPurger) run() {
	defer b.wg.Done()

	batch := make([]string, 0, maxPrefixesPerRequest)
	seen := make(map[string]struct{})

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		// 슬라이스 복사
		items := append([]string(nil), batch...)

		// 상태 초기화
		batch = batch[:0]
		clear(seen)

		b.wg.Add(1) // Worker 대기 카운트 추가
		go b.processBatch(items)
	}

	for {
		select {
		case prefix := <-b.queue:
			if _, exists := seen[prefix]; exists {
				continue
			}

			seen[prefix] = struct{}{}
			batch = append(batch, prefix)

			if len(batch) >= maxPrefixesPerRequest {
				flush()
				ticker.Reset(flushInterval)
			}

		case <-ticker.C:
			flush()

		case <-b.ctx.Done():
			flush()
			return
		}
	}
}

func (b *BatchPurger) processBatch(items []string) {
	defer b.wg.Done()

	select {
	case b.sem <- struct{}{}:
	case <-time.After(5 * time.Second):
		slog.Error("timeout acquiring semaphore during shutdown", "items", len(items))
		return
	}
	defer func() { <-b.sem }()

	startTime := time.Now()
	delay := baseRetryDelay

	for i := range maxAttempts {
		reqCtx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		err := b.client.PurgePrefixes(reqCtx, items)
		cancel()

		if err == nil {
			slog.Info("purge batch success", "count", len(items), "attempt", i+1, "duration", time.Since(startTime))
			return
		}

		if b.ctx.Err() != nil {
			slog.Warn("purge batch failed and context canceled, giving up", "error", err)
			return
		}

		slog.Warn("purge batch failed, retrying", "attempt", i+1, "retry_in", delay, "error", err)

		timer := time.NewTimer(delay)
		select {
		case <-b.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			delay *= 2
		}
	}

	slog.Error("purge batch failed permanently", "count", len(items), "duration", time.Since(startTime), "attempts", maxAttempts)
}

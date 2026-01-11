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
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		items := make([]string, len(batch))
		copy(items, batch)
		batch = batch[:0]

		b.wg.Add(1)
		go b.processBatch(items)
	}

	for {
		select {
		case prefix := <-b.queue:
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
	case <-b.ctx.Done():
		return
	}
	defer func() { <-b.sem }()

	startTime := time.Now()
	delay := baseRetryDelay

	for i := range maxAttempts {
		if b.ctx.Err() != nil {
			return
		}

		reqCtx, cancel := context.WithTimeout(b.ctx, reqTimeout)
		err := b.client.PurgePrefixes(reqCtx, items)
		cancel()

		if err == nil {
			slog.Info("purge batch success", "count", len(items), "attempt", i+1, "duration", time.Since(startTime))
			return
		}

		slog.Warn("purge batch failed, retrying", "attempt", i+1, "max_attempts", maxAttempts, "retry_in", delay, "error", err)

		timer := time.NewTimer(delay)
		select {
		case <-b.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			delay *= 2
		}
	}

	slog.Error("purge batch failed permanently", "count", len(items), "duration", time.Since(startTime), "attempts", maxAttempts, "sample_item", items[0])
}

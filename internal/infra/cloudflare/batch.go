// internal/pkg/cloudflare/batch.go
package cloudflare

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type BatchPurger struct {
	client    Purger
	queue     chan string
	closeChan chan struct{}
	wg        sync.WaitGroup
}

func NewBatchPurger(client Purger) *BatchPurger {
	if client == nil {
		return nil
	}
	bp := &BatchPurger{
		client:    client,
		queue:     make(chan string, 1000),
		closeChan: make(chan struct{}),
	}
	bp.start()
	return bp
}

func (b *BatchPurger) Add(url string) {
	select {
	case b.queue <- url:
	default:
	}
}

func (b *BatchPurger) start() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		var batch []string
		seen := make(map[string]bool)

		flush := func() {
			if len(batch) > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := b.client.PurgePrefixes(ctx, batch); err != nil {
					slog.Error("batch purge failed", "err", err, "count", len(batch))
				} else {
					slog.Info("batch purge success", "count", len(batch))
				}
				cancel()

				batch = nil
				seen = make(map[string]bool)
			}
		}

		for {
			select {
			case url := <-b.queue:
				if !seen[url] {
					batch = append(batch, url)
					seen[url] = true
				}

				if len(batch) >= 80 {
					flush()
					ticker.Reset(15 * time.Second)
				}

			case <-ticker.C:
				flush()

			case <-b.closeChan:
				flush()
				return
			}
		}
	}()
}

func (b *BatchPurger) Close() {
	close(b.closeChan)
	b.wg.Wait()
}

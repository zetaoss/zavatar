// infra/cloudflare/batch.go
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
)

type BatchPurger struct {
	client    Purger
	queue     chan string
	closeChan chan struct{}
	wg        sync.WaitGroup

	log *slog.Logger
}

func NewBatchPurger(client Purger, logger *slog.Logger) *BatchPurger {
	if client == nil {
		return nil
	}
	if logger == nil {
		logger = slog.Default()
	}
	bp := &BatchPurger{
		client:    client,
		queue:     make(chan string, 1000),
		closeChan: make(chan struct{}),
		log:       logger.With("component", "cloudflare_batch_purger"),
	}
	bp.start()
	return bp
}

func (b *BatchPurger) Add(url string) {
	select {
	case b.queue <- url:
	default:
		b.log.Warn("queue full, dropping url", "url", url)
	}
}

func (b *BatchPurger) start() {
	b.wg.Go(func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		var batch []string
		seen := make(map[string]bool)

		flush := func() {
			if len(batch) == 0 {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
			err := b.client.PurgePrefixes(ctx, batch)
			cancel()

			if err != nil {
				b.log.Error("batch purge failed", "err", err, "count", len(batch))
			} else {
				b.log.Info("batch purge success", "count", len(batch))
			}

			batch = nil
			seen = make(map[string]bool)
		}

		for {
			select {
			case url := <-b.queue:
				if !seen[url] {
					batch = append(batch, url)
					seen[url] = true
				}

				if len(batch) >= maxPrefixesPerRequest {
					flush()
					ticker.Reset(flushInterval)
				}

			case <-ticker.C:
				flush()

			case <-b.closeChan:
				flush()
				return
			}
		}
	})
}

func (b *BatchPurger) Close() {
	close(b.closeChan)
	b.wg.Wait()
}

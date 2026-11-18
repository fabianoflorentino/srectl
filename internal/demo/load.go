package demo

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// LoadConfig configures the load generator.
type LoadConfig struct {
	URL      string
	Duration time.Duration
	RPS      int
	Client   *http.Client
}

// LoadResult summarises a load run.
type LoadResult struct {
	Requests int
	Errors   int
}

// loadPaths rotates through the api-a endpoints so traffic produces a mix of
// successful, slow and erroring requests.
var loadPaths = []string{"/orders", "/orders", "/slow", "/error"}

// Load generates traffic against cfg.URL for cfg.Duration, sending roughly
// cfg.RPS requests per second. 5xx responses and transport errors count as
// errors (expected here, since /error always fails).
func Load(ctx context.Context, cfg LoadConfig) (LoadResult, error) {
	if cfg.Client == nil {
		return LoadResult{}, errors.New("load: client is required")
	}
	if cfg.RPS <= 0 {
		cfg.RPS = 1
	}
	if cfg.RPS > 1000 {
		cfg.RPS = 1000
	}
	if cfg.Duration <= 0 {
		cfg.Duration = time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Duration)
	defer cancel()

	var (
		requests atomic.Int64
		errs     atomic.Int64
		idx      atomic.Int64
		wg       sync.WaitGroup
	)

	doOne := func() {
		defer wg.Done()
		path := loadPaths[int(idx.Add(1)-1)%len(loadPaths)]
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL+path, nil)
		if err != nil {
			errs.Add(1)
			return
		}
		resp, err := cfg.Client.Do(req)
		requests.Add(1)
		if err != nil {
			errs.Add(1)
			return
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= http.StatusInternalServerError {
			errs.Add(1)
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(cfg.RPS))
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			wg.Add(1)
			go doOne()
		}
	}

	wg.Wait()
	return LoadResult{Requests: int(requests.Load()), Errors: int(errs.Load())}, nil
}

package nts

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/beevik/nts"
	"github.com/syamsv/Advtrix/config"
	"go.uber.org/zap"
)

var (
	maxInitRetries = 5
	initRetryDelay = 2 * time.Second
	checkInterval  = 10 * time.Second
	driftThreshold = 10 * time.Millisecond
)

var (
	session       *nts.Session
	baseOffset    time.Duration
	baseTime      time.Time
	monotonicBase time.Duration
	mu            sync.RWMutex
	stopCh        chan struct{}
	healthy       atomic.Bool
	log           *zap.Logger
)

// Init establishes an NTS session with the server specified in config.NTS_SERVER,
// queries the initial clock offset, and starts a background drift-correction loop.
// Retries up to 5 times on initial connection failure before calling Fatal.
func Init() {
	log = zap.L().Named("nts")
	stopCh = make(chan struct{})

	server := config.NTS_SERVER

	var err error
	for attempt := 1; attempt <= maxInitRetries; attempt++ {
		session, err = nts.NewSession(server)
		if err == nil {
			break
		}
		log.Warn("NTS session failed, retrying",
			zap.String("server", server),
			zap.Int("attempt", attempt),
			zap.Int("max", maxInitRetries),
			zap.Error(err),
		)
		if attempt < maxInitRetries {
			time.Sleep(initRetryDelay)
		}
	}
	if err != nil {
		log.Fatal("failed to establish NTS session after retries",
			zap.String("server", server),
			zap.Int("attempts", maxInitRetries),
			zap.Error(err),
		)
	}

	log.Info("NTS session established", zap.String("server", server))

	if !syncClock() {
		log.Fatal("failed initial clock sync", zap.String("server", server))
	}

	healthy.Store(true)
	go driftCorrector()
}

// Shutdown stops the background drift-correction goroutine.
func Shutdown() {
	close(stopCh)
	healthy.Store(false)
	log.Info("shutdown")
}

// Healthy returns true if the NTS session is active and the last sync succeeded.
func Healthy() bool {
	return healthy.Load()
}

// Now returns the current time corrected by the NTS-derived offset and
// adjusted for monotonic clock elapsed time since the last sync.
func Now() time.Time {
	mu.RLock()
	offset := baseOffset
	base := baseTime
	monoBase := monotonicBase
	mu.RUnlock()

	elapsed := monotonic() - monoBase
	return base.Add(elapsed + offset)
}

// Offset returns the current clock offset from the NTS server.
func Offset() time.Duration {
	mu.RLock()
	defer mu.RUnlock()
	return baseOffset
}

// syncClock queries the NTS server and captures the offset along with a
// monotonic timestamp so drift can be detected between syncs.
// Returns true on success.
func syncClock() bool {
	resp, err := session.Query()
	if err != nil {
		log.Error("NTS query failed", zap.Error(err))
		return false
	}

	if err := resp.Validate(); err != nil {
		log.Error("NTS response validation failed", zap.Error(err))
		return false
	}

	mu.Lock()
	baseOffset = resp.ClockOffset
	baseTime = time.Now()
	monotonicBase = monotonic()
	mu.Unlock()

	log.Debug("clock synced",
		zap.Duration("offset", resp.ClockOffset),
		zap.Duration("rtt", resp.RTT),
		zap.Duration("root_distance", resp.RootDistance),
	)

	return true
}

// driftCorrector periodically re-syncs with the NTS server to detect and
// correct monotonic clock drift.
func driftCorrector() {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	consecutiveFailures := 0
	const maxConsecutiveFailures = 10

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			resp, err := session.Query()
			if err != nil {
				consecutiveFailures++
				log.Warn("drift check: query failed",
					zap.Error(err),
					zap.Int("consecutive_failures", consecutiveFailures),
				)
				if consecutiveFailures >= maxConsecutiveFailures {
					healthy.Store(false)
					log.Error("NTS marked unhealthy after consecutive failures",
						zap.Int("failures", consecutiveFailures),
					)
				}
				continue
			}

			if err := resp.Validate(); err != nil {
				consecutiveFailures++
				log.Warn("drift check: validation failed", zap.Error(err))
				continue
			}

			// Reset failure counter on success.
			if consecutiveFailures > 0 {
				consecutiveFailures = 0
				healthy.Store(true)
				log.Info("NTS recovered, marked healthy")
			}

			mu.RLock()
			currentOffset := baseOffset
			mu.RUnlock()

			drift := absDuration(resp.ClockOffset - currentOffset)
			if drift > driftThreshold {
				log.Warn("clock drift detected, correcting",
					zap.Duration("drift", drift),
					zap.Duration("old_offset", currentOffset),
					zap.Duration("new_offset", resp.ClockOffset),
				)

				mu.Lock()
				baseOffset = resp.ClockOffset
				baseTime = time.Now()
				monotonicBase = monotonic()
				mu.Unlock()
			}
		}
	}
}

// monotonic returns a monotonic duration since process start, immune to
// wall-clock adjustments (NTP jumps, DST, etc.).
func monotonic() time.Duration {
	return time.Since(time.Time{})
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

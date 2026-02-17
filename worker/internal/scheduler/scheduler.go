package scheduler

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/robfig/cron/v3"
    "go.uber.org/zap"
)

// Scheduler manages scheduled scraping jobs
type Scheduler struct {
    cron      *cron.Cron
    service   ScraperService
    redis     *redis.Client
    workerID  string
    logger    *zap.Logger
    jobs      map[string]cron.EntryID
}

// ScraperService defines the interface for scraping operations
type ScraperService interface {
    ScrapeWebsite(ctx context.Context, siteName, url string) error
}

// New creates a new scheduler
func New(service ScraperService, redis *redis.Client, workerID string, logger *zap.Logger) *Scheduler {
    c := cron.New(
        cron.WithSeconds(),
        cron.WithLogger(newCronLogger(logger)),
        cron.WithChain(
            cron.SkipIfStillRunning(newCronLogger(logger)),
            cron.Recover(newCronLogger(logger)),
        ),
    )

    return &Scheduler{
        cron:     c,
        service:  service,
        redis:    redis,
        workerID: workerID,
        logger:   logger,
        jobs:     make(map[string]cron.EntryID),
    }
}

// AddJob adds a new scheduled job
func (s *Scheduler) AddJob(siteName, schedule, url string) error {
    entryID, err := s.cron.AddFunc(schedule, func() {
        s.executeJob(siteName, url)
    })

    if err != nil {
        return fmt.Errorf("adding job for %s: %w", siteName, err)
    }

    s.jobs[siteName] = entryID

    s.logger.Info("job scheduled",
        zap.String("site", siteName),
        zap.String("schedule", schedule),
        zap.String("url", url))

    return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
    s.cron.Start()
    s.logger.Info("scheduler started",
        zap.Int("jobs", len(s.jobs)))
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop(ctx context.Context) {
    s.logger.Info("stopping scheduler")
    stopCtx := s.cron.Stop()
    
    // Wait for all jobs to complete or context timeout
    select {
    case <-stopCtx.Done():
        s.logger.Info("all jobs completed")
    case <-ctx.Done():
        s.logger.Warn("scheduler stop timeout, forcing shutdown")
    }
}

// executeJob executes a scraping job with distributed locking
func (s *Scheduler) executeJob(siteName, url string) {
    ctx := context.Background()
    lockKey := fmt.Sprintf("job:lock:%s", siteName)
    lockExpiry := 10 * time.Minute

    s.logger.Info("job triggered",
        zap.String("site", siteName),
        zap.String("worker", s.workerID))

    // Try to acquire lock
    acquired, err := s.acquireLock(ctx, lockKey, lockExpiry)
    if err != nil {
        s.logger.Error("failed to acquire lock",
            zap.String("site", siteName),
            zap.Error(err))
        return
    }

    if !acquired {
        s.logger.Info("job already running on another worker",
            zap.String("site", siteName))
        return
    }

    // Ensure lock is released
    defer func() {
        if err := s.releaseLock(ctx, lockKey); err != nil {
            s.logger.Error("failed to release lock",
                zap.String("site", siteName),
                zap.Error(err))
        }
    }()

    // Execute scraping
    s.logger.Info("job started",
        zap.String("site", siteName),
        zap.String("worker", s.workerID))

    startTime := time.Now()

    err = s.service.ScrapeWebsite(ctx, siteName, url)
    duration := time.Since(startTime)

    if err != nil {
        s.logger.Error("job failed",
            zap.String("site", siteName),
            zap.Duration("duration", duration),
            zap.Error(err))
        return
    }

    s.logger.Info("job completed",
        zap.String("site", siteName),
        zap.Duration("duration", duration))
}

// acquireLock acquires a distributed lock
func (s *Scheduler) acquireLock(ctx context.Context, key string, expiry time.Duration) (bool, error) {
    result, err := s.redis.SetNX(ctx, key, s.workerID, expiry).Result()
    if err != nil {
        return false, fmt.Errorf("acquiring lock: %w", err)
    }
    return result, nil
}

// releaseLock releases a distributed lock
func (s *Scheduler) releaseLock(ctx context.Context, key string) error {
    // Only release if this worker holds the lock
    val, err := s.redis.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil // Lock already released
    }
    if err != nil {
        return fmt.Errorf("checking lock: %w", err)
    }

    if val == s.workerID {
        if err := s.redis.Del(ctx, key).Err(); err != nil {
            return fmt.Errorf("releasing lock: %w", err)
        }
    }

    return nil
}

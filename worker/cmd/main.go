package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alwanly/Houses-Prices/worker/internal/api"
	"github.com/Alwanly/Houses-Prices/worker/internal/config"
	"github.com/Alwanly/Houses-Prices/worker/internal/notification"
	"github.com/Alwanly/Houses-Prices/worker/internal/pkg/logger"
	"github.com/Alwanly/Houses-Prices/worker/internal/scheduler"
	"github.com/Alwanly/Houses-Prices/worker/internal/scrape/site"
	"github.com/Alwanly/Houses-Prices/worker/internal/service"
	"github.com/Alwanly/Houses-Prices/worker/internal/storage"
	"go.uber.org/zap"
)

func main() {
	cfgPath := flag.String("config", "./configs/config.yaml", "path to config file")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	ctx := context.Background()

	// MongoDB
	mongoDB, err := storage.NewMongoDB(ctx, cfg.MongoDB.URI, cfg.MongoDB.Database, time.Duration(cfg.MongoDB.Timeout)*time.Second)
	if err != nil {
		log.Fatal("mongodb init failed", zap.Error(err))
	}
	defer mongoDB.Close(ctx)

	// Redis
	redisWrap, err := storage.NewRedis(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("redis init failed", zap.Error(err))
	}
	defer redisWrap.Close()

	// Repository
	repo := storage.NewListingRepository(mongoDB.Database())

	// Notifier
	note := notification.NewNotifier(redisWrap.Client(), log)

	// Service
	svc := service.NewScraperService(repo, note, log)

	// Register site-specific scrapers
	for _, s := range cfg.Sites {
		if !s.Enabled {
			continue
		}
		switch s.Name {
		case "rumah123":
			r := site.NewRumah123Scraper(&s, log)
			svc.RegisterScraper(s.Name, r)
		default:
			log.Warn("no scraper for site", zap.String("site", s.Name))
		}
	}

	// Scheduler
	sched := scheduler.New(svc, redisWrap.Client(), hostnameOrPID(), log)
	for _, s := range cfg.Sites {
		if !s.Enabled {
			continue
		}
		if err := sched.AddJob(s.Name, s.Schedule, s.BaseURL); err != nil {
			log.Warn("failed to add job", zap.String("site", s.Name), zap.Error(err))
		}
	}

	sched.Start()

	// API server
	apiSrv := api.NewServer(&cfg.Server, svc, note, log)
	if err := apiSrv.Start(); err != nil {
		log.Fatal("failed to start api server", zap.Error(err))
	}

	// Wait for signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := apiSrv.Stop(shutdownCtx); err != nil {
		log.Warn("api shutdown error", zap.Error(err))
	}

	sched.Stop(shutdownCtx)

	if err := mongoDB.Close(shutdownCtx); err != nil {
		log.Warn("mongodb close error", zap.Error(err))
	}

	if err := redisWrap.Close(); err != nil {
		log.Warn("redis close error", zap.Error(err))
	}

	log.Info("shutdown complete")
}

func hostnameOrPID() string {
	hn, err := os.Hostname()
	if err == nil && hn != "" {
		return hn
	}
	return fmt.Sprintf("pid-%d", os.Getpid())
}

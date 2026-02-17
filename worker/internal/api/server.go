package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Alwanly/Houses-Prices/worker/internal/config"
	"github.com/Alwanly/Houses-Prices/worker/internal/notification"
	"github.com/Alwanly/Houses-Prices/worker/internal/service"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	svc        *service.ScraperService
	notifier   *notification.Notifier
	logger     *zap.Logger
	cfg        *config.ServerConfig
}

func NewServer(cfg *config.ServerConfig, svc *service.ScraperService, notifier *notification.Notifier, logger *zap.Logger) *Server {
	mux := http.NewServeMux()
	s := &Server{
		svc:      svc,
		notifier: notifier,
		logger:   logger,
		cfg:      cfg,
	}

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/listings", s.handleList)
	mux.HandleFunc("/scrape", s.handleTrigger)

	s.httpServer = &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", zap.String("addr", s.httpServer.Addr))
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("http server stopped", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}

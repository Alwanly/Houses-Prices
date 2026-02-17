package site

import (
    "github.com/Alwanly/Houses-Prices/worker/internal/config"
    "github.com/Alwanly/Houses-Prices/worker/internal/scrape"
    "go.uber.org/zap"
)

// BaseScraper provides common functionality for site-specific scrapers
type BaseScraper struct {
    Config *config.SiteConfig
    Colly  *scrape.CollyScraper
    Logger *zap.Logger
}

// NewBaseScraper creates a new base scraper
func NewBaseScraper(cfg *config.SiteConfig, logger *zap.Logger) *BaseScraper {
    return &BaseScraper{
        Config: cfg,
        Colly:  scrape.NewCollyScraper(cfg, logger),
        Logger: logger,
    }
}

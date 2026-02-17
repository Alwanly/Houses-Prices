package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Alwanly/Houses-Prices/worker/internal/model"
	"github.com/Alwanly/Houses-Prices/worker/internal/storage"
)

// ScraperService orchestrates scraping operations
type ScraperService struct {
	scrapers   map[string]Scraper
	repository storage.ListingRepository
	notifier   Notifier
	logger     *zap.Logger
}

// Scraper interface for site-specific scrapers
type Scraper interface {
	Scrape(ctx context.Context, url string) (*model.ScrapeResult, error)
}

// Notifier interface for notifications
type Notifier interface {
	NotifyError(ctx context.Context, siteName string, err error) error
	NotifySuccess(ctx context.Context, siteName string, count int) error
}

// NewScraperService creates a new scraper service
func NewScraperService(
	repository storage.ListingRepository,
	notifier Notifier,
	logger *zap.Logger,
) *ScraperService {
	return &ScraperService{
		scrapers:   make(map[string]Scraper),
		repository: repository,
		notifier:   notifier,
		logger:     logger,
	}
}

// RegisterScraper registers a site-specific scraper
func (s *ScraperService) RegisterScraper(name string, scraper Scraper) {
	s.scrapers[name] = scraper
	s.logger.Info("scraper registered", zap.String("site", name))
}

// ScrapeWebsite performs complete scraping workflow for a site
func (s *ScraperService) ScrapeWebsite(ctx context.Context, siteName, url string) error {
	scraper, ok := s.scrapers[siteName]
	if !ok {
		return fmt.Errorf("scraper not found for site: %s", siteName)
	}

	s.logger.Info("starting scrape job", zap.String("site", siteName), zap.String("url", url))

	// Scrape
	result, err := scraper.Scrape(ctx, url)
	if err != nil {
		if s.notifier != nil {
			s.notifier.NotifyError(ctx, siteName, err)
		}
		return fmt.Errorf("scraping %s: %w", siteName, err)
	}

	// Save each listing
	savedCount := 0
	for _, listing := range result.Listings {
		listing.SiteName = siteName

		if err := s.repository.Save(ctx, listing); err != nil {
			s.logger.Error("failed to save listing", zap.String("url", listing.URL), zap.Error(err))
			continue
		}
		savedCount++
	}

	s.logger.Info("scrape job completed",
		zap.String("site", siteName),
		zap.Int("scraped", len(result.Listings)),
		zap.Int("saved", savedCount),
		zap.Int("errors", result.ErrorCount))

	// Notify success
	if s.notifier != nil {
		s.notifier.NotifySuccess(ctx, siteName, savedCount)
	}

	return nil
}

// GetListings retrieves listings with filters
func (s *ScraperService) GetListings(ctx context.Context, filter *storage.ListingFilter) ([]*model.Listing, error) {
	return s.repository.FindAll(ctx, filter)
}

package site

import (
    "context"
    "fmt"

    "go.uber.org/zap"

    "github.com/Alwanly/Houses-Prices/worker/internal/config"
    "github.com/Alwanly/Houses-Prices/worker/internal/model"
)

// Rumah123Scraper implements scraping for rumah123.com
type Rumah123Scraper struct {
    *BaseScraper
}

// NewRumah123Scraper creates a new Rumah123 scraper
func NewRumah123Scraper(cfg *config.SiteConfig, logger *zap.Logger) *Rumah123Scraper {
    return &Rumah123Scraper{
        BaseScraper: NewBaseScraper(cfg, logger),
    }
}

// Scrape implements the Scraper interface for rumah123.com
func (s *Rumah123Scraper) Scrape(ctx context.Context, url string) (*model.ScrapeResult, error) {
    s.Logger.Info("starting rumah123 scrape",
        zap.String("url", url))

    // Use base Colly scraper
    result, err := s.Colly.Scrape(ctx, url)
    if err != nil {
        return nil, fmt.Errorf("scraping rumah123: %w", err)
    }

    // Site-specific post-processing can be added here
    // For example: CAPTCHA detection, additional data enrichment, etc.

    // Check for CAPTCHA indicators
    if s.detectCaptcha(result) {
        s.Logger.Warn("CAPTCHA detected on rumah123",
            zap.String("url", url))
        return nil, fmt.Errorf("CAPTCHA challenge encountered")
    }

    // Validate results
    if len(result.Listings) == 0 {
        s.Logger.Warn("no listings found",
            zap.String("url", url))
    }

    return result, nil
}

// detectCaptcha checks if CAPTCHA is present in results
func (s *Rumah123Scraper) detectCaptcha(result *model.ScrapeResult) bool {
    // CAPTCHA usually results in zero listings and specific error patterns
    if len(result.Listings) == 0 && len(result.Errors) > 0 {
        for _, errMsg := range result.Errors {
            // Check for common CAPTCHA-related errors
            if len(errMsg) > 0 {
                // This is a simple check - enhance based on actual site behavior
                return result.TotalScraped == 0
            }
        }
    }
    return false
}

package scrape

import (
	"context"

	"github.com/Alwanly/Houses-Prices/worker/internal/model"
)

// Scraper defines the interface for web scraping operations
type Scraper interface {
	// Scrape fetches and parses listings from the given URL
	Scrape(ctx context.Context, url string) (*model.ScrapeResult, error)
}

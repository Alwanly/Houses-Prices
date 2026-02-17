package scrape

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"

	"github.com/Alwanly/Houses-Prices/worker/internal/config"
	"github.com/Alwanly/Houses-Prices/worker/internal/model"
	"github.com/Alwanly/Houses-Prices/worker/internal/pkg/retry"
)

// CollyScraper implements Scraper using Colly framework
type CollyScraper struct {
	config    *config.SiteConfig
	collector *colly.Collector
	logger    *zap.Logger
}

// NewCollyScraper creates a new Colly-based scraper
func NewCollyScraper(cfg *config.SiteConfig, logger *zap.Logger) *CollyScraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.Async(true),
	)

	// Set rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: cfg.RateLimit,
		Delay:       time.Second / time.Duration(cfg.RateLimit),
	})

	// Set timeout
	c.SetRequestTimeout(time.Duration(cfg.Timeout) * time.Second)

	// Error handler
	c.OnError(func(r *colly.Response, err error) {
		logger.Error("scraping error",
			zap.String("url", r.Request.URL.String()),
			zap.Int("status", r.StatusCode),
			zap.Error(err))
	})

	// Request handler for logging
	c.OnRequest(func(r *colly.Request) {
		logger.Debug("visiting",
			zap.String("url", r.URL.String()))
	})

	return &CollyScraper{
		config:    cfg,
		collector: c,
		logger:    logger,
	}
}

// Scrape implements the Scraper interface
func (s *CollyScraper) Scrape(ctx context.Context, url string) (*model.ScrapeResult, error) {
	startTime := time.Now()

	result := &model.ScrapeResult{
		SiteName: s.config.Name,
		URL:      url,
		Listings: make([]*model.Listing, 0),
		Errors:   make([]string, 0),
	}

	// Clone collector for this scrape
	c := s.collector.Clone()

	// Extract listings
	c.OnHTML(s.config.Selectors.ListItem, func(e *colly.HTMLElement) {
		listing, err := s.extractListing(e)
		if err != nil {
			s.logger.Warn("failed to extract listing",
				zap.Error(err))
			result.Errors = append(result.Errors, err.Error())
			result.ErrorCount++
			return
		}

		if listing != nil {
			result.Listings = append(result.Listings, listing)
			result.TotalScraped++
		}
	})

	// Extract next page URL
	c.OnHTML(s.config.Selectors.NextPage, func(e *colly.HTMLElement) {
		nextURL := e.Attr("href")
		if nextURL != "" {
			result.NextPageURL = MakeAbsoluteURL(url, nextURL)
			result.HasNextPage = true
		}
	})

	// Visit with retry
	retryConfig := retry.DefaultConfig()
	err := retry.Do(ctx, retryConfig, func() error {
		return c.Visit(url)
	})

	if err != nil {
		return nil, fmt.Errorf("visiting url: %w", err)
	}

	// Wait for async operations
	c.Wait()

	result.Duration = time.Since(startTime).Seconds()
	result.TotalFound = result.TotalScraped

	s.logger.Info("scraping completed",
		zap.String("site", s.config.Name),
		zap.String("url", url),
		zap.Int("scraped", result.TotalScraped),
		zap.Int("errors", result.ErrorCount),
		zap.Float64("duration", result.Duration))

	return result, nil
}

// extractListing extracts a single listing from HTML element
func (s *CollyScraper) extractListing(e *colly.HTMLElement) (*model.Listing, error) {
	sel := s.config.Selectors

	// Extract required fields
	title := CleanText(e.ChildText(sel.Title))
	if title == "" {
		return nil, fmt.Errorf("missing title")
	}

	priceText := CleanText(e.ChildText(sel.Price))
	price, err := ParsePrice(priceText)
	if err != nil {
		return nil, fmt.Errorf("parsing price: %w", err)
	}

	location := CleanText(e.ChildText(sel.Location))
	if location == "" {
		return nil, fmt.Errorf("missing location")
	}

	detailURL := e.ChildAttr(sel.DetailURL, "href")
	if detailURL == "" {
		detailURL = e.Request.AbsoluteURL(e.Attr("href"))
	} else {
		detailURL = MakeAbsoluteURL(e.Request.URL.String(), detailURL)
	}

	if detailURL == "" {
		return nil, fmt.Errorf("missing detail URL")
	}

	// Extract optional fields
	bedrooms := 0
	if sel.Bedrooms != "" {
		bedrooms = ParseInt(e.ChildText(sel.Bedrooms))
	}

	bathrooms := 0
	if sel.Bathrooms != "" {
		bathrooms = ParseInt(e.ChildText(sel.Bathrooms))
	}

	landArea := 0.0
	if sel.LandArea != "" {
		landArea = ParseFloat(e.ChildText(sel.LandArea))
	}

	buildingArea := 0.0
	if sel.BuildingArea != "" {
		buildingArea = ParseFloat(e.ChildText(sel.BuildingArea))
	}

	description := ""
	if sel.Description != "" {
		description = CleanText(e.ChildText(sel.Description))
	}

	agentName := ""
	if sel.AgentName != "" {
		agentName = CleanText(e.ChildText(sel.AgentName))
	}

	agentPhone := ""
	if sel.AgentPhone != "" {
		agentPhone = CleanText(e.ChildText(sel.AgentPhone))
	}

	// Extract images
	images := make([]string, 0)
	if sel.Images != "" {
		e.ForEach(sel.Images, func(_ int, img *colly.HTMLElement) {
			src := img.Attr("src")
			if src == "" {
				src = img.Attr("data-src")
			}
			if src != "" {
				images = append(images, MakeAbsoluteURL(e.Request.URL.String(), src))
			}
		})
	}

	listing := &model.Listing{
		SiteName:     s.config.Name,
		URL:          detailURL,
		Title:        title,
		Price:        price,
		Location:     location,
		Bedrooms:     bedrooms,
		Bathrooms:    bathrooms,
		LandArea:     landArea,
		BuildingArea: buildingArea,
		Description:  description,
		Images:       images,
		AgentName:    agentName,
		AgentPhone:   agentPhone,
		ScrapedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return listing, nil
}

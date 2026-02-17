package model

// ScrapeResult represents the result of a scraping operation
type ScrapeResult struct {
	SiteName     string     `json:"site_name"`
	URL          string     `json:"url"`
	Listings     []*Listing `json:"listings"`
	TotalFound   int        `json:"total_found"`
	TotalScraped int        `json:"total_scraped"`
	ErrorCount   int        `json:"error_count"`
	Errors       []string   `json:"errors,omitempty"`
	Duration     float64    `json:"duration_seconds"`
	NextPageURL  string     `json:"next_page_url,omitempty"`
	HasNextPage  bool       `json:"has_next_page"`
}

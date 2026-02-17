package storage

import (
    "context"

    "github.com/Alwanly/Houses-Prices/worker/internal/model"
)

// ListingRepository defines operations for listing storage
type ListingRepository interface {
    Save(ctx context.Context, listing *model.Listing) error
    FindByURL(ctx context.Context, url string) (*model.Listing, error)
    FindAll(ctx context.Context, filter *ListingFilter) ([]*model.Listing, error)
    UpdatePrice(ctx context.Context, url string, newPrice float64) error
    Count(ctx context.Context, filter *ListingFilter) (int64, error)
}

// ListingFilter defines filter options for querying listings
type ListingFilter struct {
    SiteName     string
    MinPrice     float64
    MaxPrice     float64
    Location     string
    MinBedrooms  int
    MinBathrooms int
    Limit        int
    Offset       int
}

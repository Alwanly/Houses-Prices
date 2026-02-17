package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Alwanly/Houses-Prices/worker/internal/model"
	"github.com/Alwanly/Houses-Prices/worker/internal/storage"
	"go.uber.org/zap"
)

type mockRepo struct {
	saved []*model.Listing
}

func (m *mockRepo) Save(ctx context.Context, listing *model.Listing) error {
	m.saved = append(m.saved, listing)
	return nil
}

func (m *mockRepo) FindByURL(ctx context.Context, url string) (*model.Listing, error) {
	return nil, nil
}

func (m *mockRepo) FindAll(ctx context.Context, f *storage.ListingFilter) ([]*model.Listing, error) {
	return m.saved, nil
}

func (m *mockRepo) UpdatePrice(ctx context.Context, url string, newPrice float64) error {
	return nil
}

func (m *mockRepo) Count(ctx context.Context, f *storage.ListingFilter) (int64, error) {
	return int64(len(m.saved)), nil
}

type mockNotifier struct {
	lastErrSite  string
	lastErr      error
	lastSuccess  string
	lastSuccessN int
}

func (n *mockNotifier) NotifyError(ctx context.Context, siteName string, err error) error {
	n.lastErrSite = siteName
	n.lastErr = err
	return nil
}

func (n *mockNotifier) NotifySuccess(ctx context.Context, siteName string, count int) error {
	n.lastSuccess = siteName
	n.lastSuccessN = count
	return nil
}

type fakeScraperSuccess struct{}

func (f *fakeScraperSuccess) Scrape(ctx context.Context, url string) (*model.ScrapeResult, error) {
	return &model.ScrapeResult{
		Listings:     []*model.Listing{{URL: url, Title: "Test", Price: 123.45}},
		TotalScraped: 1,
	}, nil
}

type fakeScraperError struct{}

func (f *fakeScraperError) Scrape(ctx context.Context, url string) (*model.ScrapeResult, error) {
	return nil, errors.New("scrape failed")
}

func TestScrapeWebsite_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	notifier := &mockNotifier{}
	svc := NewScraperService(repo, notifier, zap.NewNop())

	svc.RegisterScraper("testsite", &fakeScraperSuccess{})

	if err := svc.ScrapeWebsite(ctx, "testsite", "http://example.com"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved listing, got %d", len(repo.saved))
	}

	if notifier.lastSuccessN != 1 {
		t.Fatalf("expected notifier success count 1, got %d", notifier.lastSuccessN)
	}
}

func TestScrapeWebsite_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	notifier := &mockNotifier{}
	svc := NewScraperService(repo, notifier, zap.NewNop())

	if err := svc.ScrapeWebsite(ctx, "nosite", "http://example.com"); err == nil {
		t.Fatalf("expected error for missing scraper, got nil")
	}
}

func TestScrapeWebsite_ScrapeError_NotifierCalled(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	notifier := &mockNotifier{}
	svc := NewScraperService(repo, notifier, zap.NewNop())

	svc.RegisterScraper("badsite", &fakeScraperError{})

	if err := svc.ScrapeWebsite(ctx, "badsite", "http://example.com"); err == nil {
		t.Fatalf("expected error from scraper, got nil")
	}

	if notifier.lastErr == nil {
		t.Fatalf("expected notifier to be called with error")
	}
}

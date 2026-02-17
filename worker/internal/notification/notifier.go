package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Notifier publishes notifications to Redis pub/sub
type Notifier struct {
	redis  *redis.Client
	logger *zap.Logger
}

// ErrorNotification represents an error notification
type ErrorNotification struct {
	Type      string    `json:"type"`
	SiteName  string    `json:"site_name"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// SuccessNotification represents a success notification
type SuccessNotification struct {
	Type      string    `json:"type"`
	SiteName  string    `json:"site_name"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// NewNotifier creates a new notifier
func NewNotifier(redis *redis.Client, logger *zap.Logger) *Notifier {
	return &Notifier{
		redis:  redis,
		logger: logger,
	}
}

// NotifyError publishes an error notification
func (n *Notifier) NotifyError(ctx context.Context, siteName string, err error) error {
	notification := ErrorNotification{
		Type:      "error",
		SiteName:  siteName,
		Error:     err.Error(),
		Timestamp: time.Now(),
	}

	data, marshalErr := json.Marshal(notification)
	if marshalErr != nil {
		return fmt.Errorf("marshaling error notification: %w", marshalErr)
	}

	if err := n.redis.Publish(ctx, "scraper:notifications", string(data)).Err(); err != nil {
		return fmt.Errorf("publishing error notification: %w", err)
	}

	n.logger.Info("error notification sent",
		zap.String("site", siteName),
		zap.String("error", err.Error()))

	return nil
}

// NotifySuccess publishes a success notification
func (n *Notifier) NotifySuccess(ctx context.Context, siteName string, count int) error {
	notification := SuccessNotification{
		Type:      "success",
		SiteName:  siteName,
		Count:     count,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("marshaling success notification: %w", err)
	}

	if err := n.redis.Publish(ctx, "scraper:notifications", string(data)).Err(); err != nil {
		return fmt.Errorf("publishing success notification: %w", err)
	}

	n.logger.Info("success notification sent",
		zap.String("site", siteName),
		zap.Int("count", count))

	return nil
}

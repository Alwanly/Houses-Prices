package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB wraps MongoDB client and database
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(ctx context.Context, uri, database string, timeout time.Duration) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set client options
	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(timeout).
		SetConnectTimeout(timeout)

	// Connect
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connecting to mongodb: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}

	return &MongoDB{
		client:   client,
		database: client.Database(database),
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnecting from mongodb: %w", err)
	}
	return nil
}

// Database returns the database handle
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Collection returns a collection handle
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

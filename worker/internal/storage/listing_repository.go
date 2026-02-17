package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Alwanly/Houses-Prices/worker/internal/model"
)

type mongoListingRepository struct {
	collection *mongo.Collection
}

// NewListingRepository creates a new listing repository
func NewListingRepository(db *mongo.Database) ListingRepository {
	collection := db.Collection("listings")

	// Create indexes in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// URL index for duplicate detection
		urlIndex := mongo.IndexModel{
			Keys:    bson.M{"url": 1},
			Options: options.Index().SetUnique(true),
		}

		// Site name index for filtering
		siteIndex := mongo.IndexModel{
			Keys: bson.M{"site_name": 1},
		}

		// Price index for range queries
		priceIndex := mongo.IndexModel{
			Keys: bson.M{"price": 1},
		}

		// Scraped at index for sorting
		scrapedIndex := mongo.IndexModel{
			Keys: bson.M{"scraped_at": -1},
		}

		collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			urlIndex,
			siteIndex,
			priceIndex,
			scrapedIndex,
		})
	}()

	return &mongoListingRepository{
		collection: collection,
	}
}

func (r *mongoListingRepository) Save(ctx context.Context, listing *model.Listing) error {
	now := time.Now()

	// Check if listing exists
	existing, err := r.FindByURL(ctx, listing.URL)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing listing
		listing.CreatedAt = existing.CreatedAt
		listing.UpdatedAt = now
		listing.ScrapedAt = now
	} else {
		// New listing
		listing.CreatedAt = now
		listing.UpdatedAt = now
		listing.ScrapedAt = now
	}

	// Upsert based on URL
	filter := bson.M{"url": listing.URL}
	update := bson.M{"$set": listing}

	opts := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("saving listing: %w", err)
	}

	return nil
}

func (r *mongoListingRepository) FindByURL(ctx context.Context, url string) (*model.Listing, error) {
	var listing model.Listing

	filter := bson.M{"url": url}
	err := r.collection.FindOne(ctx, filter).Decode(&listing)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding listing: %w", err)
	}

	return &listing, nil
}

func (r *mongoListingRepository) FindAll(ctx context.Context, f *ListingFilter) ([]*model.Listing, error) {
	filter := bson.M{}

	if f != nil {
		if f.SiteName != "" {
			filter["site_name"] = f.SiteName
		}
		if f.MinPrice > 0 || f.MaxPrice > 0 {
			priceFilter := bson.M{}
			if f.MinPrice > 0 {
				priceFilter["$gte"] = f.MinPrice
			}
			if f.MaxPrice > 0 {
				priceFilter["$lte"] = f.MaxPrice
			}
			filter["price"] = priceFilter
		}
		if f.Location != "" {
			filter["location"] = bson.M{"$regex": f.Location, "$options": "i"}
		}
		if f.MinBedrooms > 0 {
			filter["bedrooms"] = bson.M{"$gte": f.MinBedrooms}
		}
		if f.MinBathrooms > 0 {
			filter["bathrooms"] = bson.M{"$gte": f.MinBathrooms}
		}
	}

	opts := options.Find().
		SetSort(bson.M{"scraped_at": -1})

	if f != nil {
		if f.Limit > 0 {
			opts.SetLimit(int64(f.Limit))
		}
		if f.Offset > 0 {
			opts.SetSkip(int64(f.Offset))
		}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("finding listings: %w", err)
	}
	defer cursor.Close(ctx)

	var listings []*model.Listing
	if err := cursor.All(ctx, &listings); err != nil {
		return nil, fmt.Errorf("decoding listings: %w", err)
	}

	return listings, nil
}

func (r *mongoListingRepository) UpdatePrice(ctx context.Context, url string, newPrice float64) error {
	filter := bson.M{"url": url}
	update := bson.M{
		"$set": bson.M{
			"price":      newPrice,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("updating price: %w", err)
	}

	return nil
}

func (r *mongoListingRepository) Count(ctx context.Context, f *ListingFilter) (int64, error) {
	filter := bson.M{}

	if f != nil {
		if f.SiteName != "" {
			filter["site_name"] = f.SiteName
		}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("counting listings: %w", err)
	}

	return count, nil
}

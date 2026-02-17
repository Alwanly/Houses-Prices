package model

import "time"

// Listing represents a house listing scraped from a website
type Listing struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	SiteName     string    `json:"site_name" bson:"site_name" validate:"required"`
	URL          string    `json:"url" bson:"url" validate:"required,url"`
	Title        string    `json:"title" bson:"title" validate:"required"`
	Price        float64   `json:"price" bson:"price" validate:"required,gt=0"`
	Location     string    `json:"location" bson:"location" validate:"required"`
	Bedrooms     int       `json:"bedrooms" bson:"bedrooms" validate:"min=0"`
	Bathrooms    int       `json:"bathrooms" bson:"bathrooms" validate:"min=0"`
	LandArea     float64   `json:"land_area" bson:"land_area" validate:"min=0"`
	BuildingArea float64   `json:"building_area" bson:"building_area" validate:"min=0"`
	Description  string    `json:"description" bson:"description"`
	Images       []string  `json:"images" bson:"images"`
	AgentName    string    `json:"agent_name,omitempty" bson:"agent_name,omitempty"`
	AgentPhone   string    `json:"agent_phone,omitempty" bson:"agent_phone,omitempty"`
	ScrapedAt    time.Time `json:"scraped_at" bson:"scraped_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}

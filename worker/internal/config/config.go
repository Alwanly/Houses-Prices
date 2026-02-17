package config

// Config represents the application configuration
type Config struct {
    Server  ServerConfig  `mapstructure:"server" validate:"required"`
    MongoDB MongoDBConfig `mapstructure:"mongodb" validate:"required"`
    Redis   RedisConfig   `mapstructure:"redis" validate:"required"`
    Logging LoggingConfig `mapstructure:"logging" validate:"required"`
    Sites   []SiteConfig  `mapstructure:"sites" validate:"required,min=1,dive"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
    Port            int `mapstructure:"port" validate:"required,min=1,max=65535"`
    ShutdownTimeout int `mapstructure:"shutdown_timeout" validate:"required,min=1"`
}

// MongoDBConfig holds MongoDB connection configuration
type MongoDBConfig struct {
    URI      string `mapstructure:"uri" validate:"required"`
    Database string `mapstructure:"database" validate:"required"`
    Timeout  int    `mapstructure:"timeout" validate:"required,min=1"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
    Addr     string `mapstructure:"addr" validate:"required"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db" validate:"min=0"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
    Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
    Format string `mapstructure:"format" validate:"required,oneof=json console"`
}

// SiteConfig holds configuration for a scraping target site
type SiteConfig struct {
    Name      string         `mapstructure:"name" validate:"required"`
    BaseURL   string         `mapstructure:"base_url" validate:"required,url"`
    Schedule  string         `mapstructure:"schedule" validate:"required"`
    Enabled   bool           `mapstructure:"enabled"`
    RateLimit int            `mapstructure:"rate_limit" validate:"min=1"`
    Timeout   int            `mapstructure:"timeout" validate:"min=1"`
    Selectors SelectorConfig `mapstructure:"selectors" validate:"required"`
}

// SelectorConfig holds CSS selectors for extracting data
type SelectorConfig struct {
    ListItem     string `mapstructure:"list_item" validate:"required"`
    Title        string `mapstructure:"title" validate:"required"`
    Price        string `mapstructure:"price" validate:"required"`
    Location     string `mapstructure:"location" validate:"required"`
    DetailURL    string `mapstructure:"detail_url" validate:"required"`
    Bedrooms     string `mapstructure:"bedrooms"`
    Bathrooms    string `mapstructure:"bathrooms"`
    LandArea     string `mapstructure:"land_area"`
    BuildingArea string `mapstructure:"building_area"`
    Description  string `mapstructure:"description"`
    Images       string `mapstructure:"images"`
    AgentName    string `mapstructure:"agent_name"`
    AgentPhone   string `mapstructure:"agent_phone"`
    NextPage     string `mapstructure:"next_page"`
}

package config

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/spf13/viper"
)

// Load loads configuration from file with environment variable overrides
func Load(path string) (*Config, error) {
    v := viper.New()

    // Set config file
    v.SetConfigFile(path)
    v.SetConfigType("yaml")

    // Enable environment variable overrides
    v.SetEnvPrefix("WORKER")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // Read config file
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }

    // Unmarshal into struct
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }

    // Validate
    validate := validator.New()
    if err := validate.Struct(&cfg); err != nil {
        return nil, fmt.Errorf("validating config: %w", err)
    }

    return &cfg, nil
}

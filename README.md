# Houses-Prices

## Overview

Houses-Prices is a Go-based platform that scrapes Indonesian real estate websites, stores property listings in MongoDB, and exposes a small HTTP API for querying and triggering scrapes. It's designed to run as a distributed worker with Redis for locking and caching. The project focuses on reliability, testability, and ease of adding new site scrapers.

## Features

- Scrapes property listings from site-specific scrapers (currently `rumah123`)
- Cron-based scheduled scraping with per-site schedules
- Duplicate detection and upsert to MongoDB
- Redis-backed distributed locking and notifications
- Simple HTTP API: health, list listings, manual scrape trigger
- Config-driven selectors and rate limits using YAML + environment overrides

## Architecture

The repository follows a layered architecture with clear separation of concerns:

- `cmd/` — application entry point
- `internal/api/` — HTTP server, handlers, and middleware
- `internal/scheduler/` — cron job orchestration and scheduling
- `internal/service/` — orchestration of scrapers and business logic
- `internal/scrape/` — colly-based scraping engine and site implementations
- `internal/storage/` — MongoDB and Redis wrappers and repositories
- `internal/notification/` — publish/subscribe for job events

## Tech Stack

- Go 1.24.x
- MongoDB (driver: `go.mongodb.org/mongo-driver`)
- Redis (client: `github.com/redis/go-redis/v9`)
- Colly (`github.com/gocolly/colly/v2`) for scraping
- Robfig Cron (`github.com/robfig/cron/v3`) for scheduling
- Viper (`github.com/spf13/viper`) for config loading
- Zap (`go.uber.org/zap`) for structured logging

## Prerequisites

- Go 1.24 or newer
- MongoDB instance (local or remote)
- Redis instance (local or remote)

## Installation

Clone the repo and build the worker:

```bash
git clone <repository-url>
cd Houses-Prices/worker
go mod download
go build -o bin/worker ./cmd/main.go
```

Run with the example configuration:

```bash
./bin/worker -config ./configs/config.yaml
# or
go run ./cmd/main.go -config ./configs/config.yaml
```

## Configuration

Configuration is YAML-based and located in `worker/configs/config.yaml`. Environment variables prefixed with `WORKER_` can override values (dots -> underscores).

Example `config.example.yaml` structure:

```yaml
server:
	port: 8080
	shutdown_timeout: 30

mongodb:
	uri: "mongodb://localhost:27017"
	database: "houses_prices"
	timeout: 10

redis:
	addr: "localhost:6379"
	password: ""
	db: 0

logging:
	level: "info"
	format: "json"

sites:
	- name: "rumah123"
		base_url: "https://www.rumah123.com/jual/jakarta-selatan/rumah/"
		schedule: "0 0 2 * * *"
		enabled: true
		rate_limit: 2
		timeout: 30
		selectors:
			list_item: ".card-featured"
			title: ".card-featured__content-title"
			price: ".card-featured__content-price"
			location: ".card-featured__content-address"
			detail_url: "a.card-featured__link"
			next_page: "a.pagination__next"
```

### Environment variable examples

```bash
export WORKER_MONGODB_URI="mongodb://user:pass@mongo:27017"
export WORKER_SERVER_PORT=9090
```

## Running Tests

Run unit tests:

```bash
cd worker
go test ./... -v
```

## HTTP API

The worker exposes a minimal HTTP API (see `internal/api/`):

- `GET /health` — returns 200 OK when the service is healthy
- `GET /listings` — paginated list of saved listings (query params supported)
- `POST /scrape?site=<site>&url=<optional_url>` — trigger manual scrape for site; if `url` is provided, scrapes that single page

Example curl calls:

```bash
curl http://localhost:8080/health

curl "http://localhost:8080/listings?limit=20&page=1"

curl -X POST "http://localhost:8080/scrape?site=rumah123"
curl -X POST "http://localhost:8080/scrape?site=rumah123&url=https://www.rumah123.com/...."
```

## Data model & Indexes

Listings are stored in MongoDB with an upsert strategy by URL. Key fields include:

- `url` (unique index)
- `site_name`
- `title`
- `price` (numeric)
- `location`
- `bedrooms`, `bathrooms`, `land_area`, `building_area`
- `images` (array)
- `scraped_at`

Indexes (implemented in `listing_repository.go`):

- unique index on `url`
- index on `site_name`
- index on `price`

## Adding a New Scraper

To add support for a new site:

1. Create `internal/scrape/site/<sitename>.go` implementing site-specific selectors and parsing
2. Register the scraper in `internal/service/scraper_service.go`
3. Add configuration entry under `sites:` in the YAML
4. Add unit tests for parsing logic and service integration

## Development

- Build: `go build ./...`
- Run tests: `go test ./... -v`
- Lint/format: use `gofmt`/`golangci-lint` as preferred

## Deployment

Recommended deployment approach:

- Containerize the worker and run with a `docker-compose.yml` including MongoDB and Redis for local development
- Use Redis locking when running multiple worker replicas to avoid duplicate jobs

## Troubleshooting

- If scrapes return no results: check CSS selectors in config for the site
- If jobs collide across workers: verify Redis connectivity and correct lock keys
- If MongoDB upserts fail: check unique index on `url` and connection URI

## Project Status

- Worker: functional (scraper for `rumah123`, scheduling, storage, notifications)
- Web frontend: placeholder directory (`web/`) — not implemented yet
- Docs: `docs/` is currently empty; this README fills initial documentation needs

## Contributing

Contributions are welcome. Please open issues or PRs. Follow standard Go project practices and include tests for new logic.

## License

This repository is licensed under the terms in the `LICENSE` file.
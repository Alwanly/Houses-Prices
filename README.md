# Houses-Prices

## Overview

Houses-Prices is a Go worker that scrapes Indonesian real estate listings, stores them in MongoDB, and provides an HTTP API for checking health, listing saved properties, and starting scrapes. Redis is used for distributed job locking and notifications.

## Features

- Scraping for supported real estate sites (currently Rumah123)
- Cron-based scheduled scraping
- Duplicate detection and upserts in MongoDB
- Redis-backed distributed locking and notifications
- HTTP endpoints for health checks, listings, and manual scrape requests
- YAML configuration with environment-variable overrides

## Tech stack

- Go 1.27.3
- MongoDB (`go.mongodb.org/mongo-driver`)
- Redis (`github.com/redis/go-redis/v9`)
- Colly, Robfig Cron, Viper, and Zap

## Prerequisites

- Go 1.27.3 or newer
- MongoDB, local or remote
- Redis, local or remote

## Installation and running

From the repository root, enter the worker module and download the Go modules:

```bash
cd worker
go mod download
```

Edit `configs/config.yaml` with the MongoDB and Redis connection details, then build and run the worker from the `worker` directory. The repository includes this configuration file; use `configs/config.example.yaml` as a reference when changing it.

```bash
go build -o bin/worker ./cmd
./bin/worker
```

The worker reads `./configs/config.yaml` by default. To use a different file, pass its path with `-config`:

```bash
./bin/worker -config ./configs/config.example.yaml
```

## Configuration

The complete example configuration is in `worker/configs/config.example.yaml`. Settings cover the HTTP server, MongoDB, Redis, logging, and each scraper site. Environment variables prefixed with `WORKER_` can override configuration values; nested keys use underscores:

```bash
export WORKER_MONGODB_URI="mongodb://localhost:27017"
export WORKER_SERVER_PORT=9090
export WORKER_REDIS_ADDR="localhost:6379"
```

The scraper schedule uses a six-field cron expression (seconds, minutes, hours, day of month, month, day of week). CSS selectors are site-specific and may need updating if the target website changes its markup.

## HTTP API

The API listens on the configured server port (8080 by default):

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Returns `200 OK` and `{"status":"ok"}`. |
| `GET` | `/listings` | Returns saved listings in an `items` array. Pagination is not currently implemented. |
| `POST` | `/scrape?site=<site>` | Starts a background scrape and returns `202 Accepted`. |

The scrape endpoint accepts an optional `url` query parameter to pass a URL to the scraper:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/listings
curl -X POST "http://localhost:8080/scrape?site=rumah123"
curl -X POST "http://localhost:8080/scrape?site=rumah123&url=https%3A%2F%2Fwww.rumah123.com%2Fjual%2Fjakarta-selatan%2Frumah%2F"
```

## Development

Run tests and build from the repository root:

```bash
(cd worker && go test ./... -v)
(cd worker && go build ./...)
```

Format Go sources with `gofmt`.

## Deployment

Deploy the worker with reachable MongoDB and Redis services. A `docker-compose.yml` is not currently included. Redis locking can be used to coordinate multiple worker replicas and avoid duplicate scheduled jobs.

## Troubleshooting

- If scrapes return no results, check the configured CSS selectors against the site's current markup.
- If scheduled jobs collide across workers, check Redis connectivity and lock configuration.
- If MongoDB upserts fail, check the connection URI and the unique index on listing URLs.

## Contributing

Contributions are welcome. Please include tests for new behavior and follow standard Go formatting and project practices.

## License

This repository is licensed under the terms in `LICENSE`.

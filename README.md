# Stockyard Mirage

**API mock server — define endpoints and response shapes, serve fake JSON instantly**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted developer tools.

## Quick Start

```bash
docker run -p 9050:9050 -v mirage_data:/data ghcr.io/stockyard-dev/stockyard-mirage
```

Or with docker-compose:

```bash
docker-compose up -d
```

Open `http://localhost:9050` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9050` | HTTP port |
| `DATA_DIR` | `./data` | SQLite database directory |
| `MIRAGE_LICENSE_KEY` | *(empty)* | Pro license key |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 10 endpoints, 500 logged requests | Unlimited endpoints and logs |
| Price | Free | $2.99/mo |

Get a Pro license at [stockyard.dev/tools/](https://stockyard.dev/tools/).

## Category

Developer Tools

## License

Apache 2.0

# CleanCast

<div align="center">

[![Contributors](https://img.shields.io/github/contributors/ikoyhn/clean-cast.svg?style=for-the-badge)](https://github.com/ikoyhn/clean-cast/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/ikoyhn/clean-cast.svg?style=for-the-badge)](https://github.com/ikoyhn/clean-cast/network/members)
[![Stargazers](https://img.shields.io/github/stars/ikoyhn/clean-cast.svg?style=for-the-badge)](https://github.com/ikoyhn/clean-cast/stargazers)
[![Issues](https://img.shields.io/github/issues/ikoyhn/clean-cast.svg?style=for-the-badge)](https://github.com/ikoyhn/clean-cast/issues)
[![MIT License](https://img.shields.io/github/license/ikoyhn/clean-cast.svg?style=for-the-badge)](https://github.com/ikoyhn/clean-cast/blob/master/LICENSE.txt)

**Podcasting, Purified**

Transform YouTube videos into ad-free, sponsor-free podcast feeds with automatic content filtering and intelligent episode management.

[Report Bug](https://github.com/ikoyhn/clean-cast/issues/new?labels=bug) • [Request Feature](https://github.com/ikoyhn/clean-cast/issues/new?labels=enhancement) • [Documentation](https://github.com/ikoyhn/clean-cast/tree/main/docs)

</div>

---

## Table of Contents

- [About](#about)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Quick Start with Docker](#quick-start-with-docker)
  - [Manual Installation](#manual-installation)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Advanced Configuration](#advanced-configuration)
- [Usage](#usage)
  - [Creating RSS Feeds](#creating-rss-feeds)
  - [API Endpoints](#api-endpoints)
- [Development](#development)
  - [Building from Source](#building-from-source)
  - [Running Tests](#running-tests)
  - [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## About

CleanCast is a self-hosted Go application that transforms YouTube content into clean, ad-free podcast feeds. It automatically removes sponsored segments using SponsorBlock data, downloads audio on-demand, and serves it through standard RSS feeds compatible with any podcast application.

### Key Highlights

- **Zero Ads**: Automatically removes sponsored segments, intros, outros, and more
- **On-Demand Downloads**: Episodes are downloaded only when requested, saving storage
- **Standard RSS**: Works with Apple Podcasts, Pocket Casts, VLC, and any RSS-compatible app
- **Smart Caching**: Intelligent episode management with automatic cleanup
- **Multi-Format Support**: M4A, MP3, and Opus audio formats
- **Analytics**: Track listening habits and popular episodes
- **Backup & Restore**: Full database and audio file backups with S3 support
- **Self-Hosted**: Complete control over your data and privacy

---

## Features

### Core Functionality
- **YouTube to RSS Conversion**: Convert any YouTube playlist or channel into an RSS podcast feed
- **Automatic Ad Removal**: Integrates with SponsorBlock to remove:
  - Sponsor segments
  - Self-promotion
  - Interaction reminders
  - Intros and outros
  - Preview/recap segments
  - Music offtopic segments
  - Filler content

### Content Management
- **Smart Playlists**: Create custom playlists based on filters and criteria
- **Content Filters**: Filter episodes by title, duration, keywords
- **OPML Support**: Import/export podcast subscriptions
- **Episode Search**: Full-text search across all episodes and podcasts
- **Batch Operations**: Add/refresh/delete multiple podcasts at once

### Advanced Features
- **Analytics Dashboard**: Track plays, geographic distribution, popular episodes
- **Transcript Support**: Access and search video transcripts
- **Webhook Integration**: Discord, Slack, and generic webhook notifications
- **Recommendations**: Get personalized podcast recommendations
- **Backup System**: Automated backups with S3 cloud storage support
- **Prometheus Metrics**: Built-in monitoring and observability
- **Health Checks**: Kubernetes-ready health and readiness endpoints

### API & Integration
- **RESTful API**: Comprehensive REST API for all features
- **Swagger Documentation**: Interactive API documentation at `/docs`
- **Rate Limiting**: Configurable rate limits per endpoint
- **Multiple Audio Formats**: Support for M4A, MP3, and Opus
- **Custom Audio Quality**: Configurable bitrate settings

---

## Architecture

CleanCast follows a clean architecture pattern with clear separation of concerns:

```
internal/
├── api/           # HTTP handlers and request validation
├── app/           # Application initialization and routing
├── cache/         # In-memory caching layer
├── config/        # Configuration management
├── database/      # Database models and repositories
├── middleware/    # HTTP middleware (auth, rate limiting, logging)
├── models/        # Domain models and DTOs
├── services/      # Business logic layer
│   ├── analytics/     # Usage tracking and reporting
│   ├── backup/        # Backup and restore operations
│   ├── channel/       # YouTube channel processing
│   ├── downloader/    # Audio download management
│   ├── filter/        # Content filtering
│   ├── playlist/      # Playlist management
│   ├── recommendations/ # ML-based recommendations
│   ├── rss/           # RSS feed generation
│   ├── search/        # Full-text search
│   ├── sponsorblock/  # Ad segment removal
│   ├── transcript/    # Transcript fetching
│   └── youtube/       # YouTube API integration
└── logger/        # Structured logging
```

### Technology Stack

- **Language**: Go 1.24
- **Web Framework**: Echo v4
- **Database**: SQLite with GORM
- **YouTube Integration**: Google YouTube Data API v3
- **Download Engine**: yt-dlp (via go-ytdlp)
- **Logging**: Zerolog
- **Metrics**: Prometheus
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker with multi-platform support

---

## Getting Started

### Prerequisites

1. **YouTube API Key**: Get your free API key from [Google Cloud Console](https://developers.google.com/youtube/v3/getting-started)
2. **Docker** (recommended) or **Go 1.24+** for manual installation
3. Storage space for audio files (varies based on usage)

### Quick Start with Docker

The fastest way to get started:

```bash
# Using Docker Run
docker run -d \
  --name clean-cast \
  -p 8080:8080 \
  -e GOOGLE_API_KEY=your_api_key_here \
  -v /path/to/audio:/config \
  ikoyhn/clean-cast:latest

# Using Docker Compose (recommended)
# 1. Copy the docker-compose.yml from this repo
# 2. Create a .env file with your configuration
cp .env.example .env
# Edit .env with your GOOGLE_API_KEY

# 3. Start the service
docker-compose up -d
```

The service will be available at `http://localhost:8080`

### Manual Installation

For development or custom deployments:

```bash
# 1. Clone the repository
git clone https://github.com/ikoyhn/clean-cast.git
cd clean-cast

# 2. Set required environment variables
export GOOGLE_API_KEY="your_api_key_here"
export CONFIG_DIR="./config"
export PORT="8080"

# 3. Build and run
go build -o clean-cast ./cmd/app
./clean-cast
```

---

## Configuration

### Environment Variables

#### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GOOGLE_API_KEY` | YouTube Data API v3 key | `AIzaSyC...` |

#### Optional Core Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PORT` | HTTP server port | `8080` | `3000` |
| `HOST` | Host to bind to | `0.0.0.0` | `localhost` |
| `CONFIG_DIR` | Configuration directory | `/config` | `/data/config` |
| `AUDIO_DIR` | Audio storage directory | `{CONFIG_DIR}/audio` | `/data/audio` |

#### Content Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SPONSORBLOCK_CATEGORIES` | Segments to remove | `sponsor` | `sponsor,intro,outro` |
| `MIN_DURATION` | Minimum episode duration | `5m` | `10m` |
| `COOKIES_FILE` | YT-DLP cookies filename | - | `cookies.txt` |
| `AUDIO_FORMAT` | Output audio format | `m4a` | `mp3`, `opus` |
| `AUDIO_QUALITY` | Audio bitrate | `192k` | `128k`, `256k` |

#### Scheduling & Maintenance

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `CRON` | Cleanup cron schedule | `0 0 * * 0` | `0 2 * * *` |
| `BACKUP_CRON` | Backup cron schedule | - | `0 3 * * *` |
| `BACKUP_INCLUDE_AUDIO` | Include audio in backups | `false` | `true` |

#### S3 Backup Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `BACKUP_S3_BUCKET` | S3 bucket name | `my-backups` |
| `BACKUP_S3_REGION` | AWS region | `us-east-1` |
| `BACKUP_S3_ACCESS_KEY` | AWS access key | `AKIA...` |
| `BACKUP_S3_SECRET_KEY` | AWS secret key | `secret...` |

### Advanced Configuration

#### SponsorBlock Categories

Available categories for `SPONSORBLOCK_CATEGORIES` (comma-separated):

- `sponsor` - Paid promotions
- `selfpromo` - Unpaid self-promotion
- `interaction` - Interaction reminders (subscribe, like)
- `intro` - Intermission/intro animation
- `outro` - Endcards/credits
- `preview` - Preview/recap of other videos
- `music_offtopic` - Non-music portions in music videos
- `poi_highlight` - Highlights of points of interest
- `filler` - Filler tangent/off-topic

#### Docker Volume Configuration

For production, use named volumes or bind mounts:

```yaml
volumes:
  audio_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /mnt/storage/clean-cast/audio
```

---

## Usage

### Creating RSS Feeds

#### From a Playlist

1. Find your playlist ID from YouTube URL:
   ```
   https://www.youtube.com/playlist?list=PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222
                                         └── This is your playlist ID
   ```

2. Build your RSS feed URL:
   ```
   http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222
   ```

3. Add to your podcast app

#### From a Channel

1. Get the channel ID using [this tool](https://www.tunepocket.com/youtube-channel-id-finder/)

2. Build your RSS feed URL:
   ```
   http://localhost:8080/channel/UCoj1ZgGoSBoonNZqMsVUfAA
   ```

3. Optional: Limit by date to save API quota:
   ```
   http://localhost:8080/channel/UCoj1ZgGoSBoonNZqMsVUfAA?date=06-01-2025
   ```

### API Endpoints

CleanCast provides a comprehensive REST API. Key endpoints include:

#### RSS Feeds
- `GET /rss/:playlistId` - Get RSS feed for a playlist
- `GET /channel/:channelId` - Get RSS feed for a channel
- `GET /media/:videoId` - Stream audio file

#### Content Management
- `POST /api/opml/import` - Import OPML file
- `GET /api/opml/export` - Export OPML file
- `POST /api/filters` - Create content filter
- `GET /api/filters` - List all filters

#### Smart Playlists
- `POST /api/playlist/smart` - Create smart playlist
- `GET /api/playlist/smart` - List smart playlists
- `GET /rss/smart/:id` - Get smart playlist RSS feed

#### Analytics
- `GET /api/analytics/popular` - Get popular episodes
- `GET /api/analytics/summary` - Get analytics summary
- `GET /api/analytics/dashboard` - Get dashboard data

#### Backup & Restore
- `POST /api/backup/create` - Create backup
- `GET /api/backup/list` - List backups
- `POST /api/backup/restore/:id` - Restore backup

#### Search
- `GET /search/episodes` - Search episodes
- `GET /search/podcasts` - Search podcasts

#### System
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics
- `GET /docs` - Swagger API documentation

For complete API documentation, visit `/docs` when running the application.

---

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/ikoyhn/clean-cast.git
cd clean-cast

# Install dependencies
go mod download

# Build
go build -o clean-cast ./cmd/app

# Run
./clean-cast
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/rss/

# Run benchmarks
go test -bench=. ./internal/services/rss/
```

### Contributing

Contributions are welcome and greatly appreciated!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please ensure your code follows Go conventions and includes appropriate tests.

---

## Troubleshooting

### Common Issues

#### YouTube API Quota Exceeded

**Problem**: Error messages about quota limits

**Solution**:
- Use the `?date=MM-DD-YYYY` parameter on channel feeds to limit the number of videos
- Use the `?limit=N` parameter to limit the number of episodes
- Consider using multiple API keys (rotate them)
- Wait for the quota to reset (daily at midnight Pacific Time)

#### Audio File Not Found

**Problem**: 404 error when trying to play an episode

**Solution**:
- Check that the audio directory is writable
- Verify sufficient disk space
- Check Docker volume mounts are correct
- Review logs for download errors

#### Slow First Play

**Problem**: Long wait time when playing an episode for the first time

**Solution**:
- This is expected - the episode is being downloaded on-demand
- Subsequent plays will be instant (cached)
- Consider pre-downloading popular episodes
- Check your internet connection speed

#### Podcast App Not Updating

**Problem**: New episodes don't appear in your podcast app

**Solution**:
- Most apps cache RSS feeds for hours
- Force refresh in your podcast app
- Check if the YouTube channel has actually posted new videos
- Verify cron job is running for metadata refresh

#### Memory Usage

**Problem**: High memory consumption

**Solution**:
- Adjust Docker memory limits in docker-compose.yml
- Clear old audio files (adjust `CRON` schedule)
- Reduce concurrent downloads
- Check for memory leaks (file an issue)

### Logging

Enable debug logging for troubleshooting:

```bash
# Set log level
export LOG_LEVEL=debug

# View Docker logs
docker-compose logs -f clean-cast

# View logs with timestamps
docker-compose logs -f --tail=100 clean-cast
```

### Getting Help

- Check the [Discussions](https://github.com/ikoyhn/clean-cast/discussions) for setup guides
- Search [existing issues](https://github.com/ikoyhn/clean-cast/issues)
- Create a [new issue](https://github.com/ikoyhn/clean-cast/issues/new) with:
  - Your configuration (redact sensitive data)
  - Relevant logs
  - Steps to reproduce
  - Expected vs actual behavior

---

## FAQ

**Q: Is this legal?**
A: CleanCast uses official YouTube APIs and yt-dlp, which is widely used. You should only use it for content you have the right to access. Check your local laws and YouTube's Terms of Service.

**Q: Does this work with private/unlisted videos?**
A: Yes, if you provide cookies via `COOKIES_FILE` from an authenticated YouTube session.

**Q: Can I use this with other video platforms?**
A: Currently only YouTube is supported. Other platforms may be added in the future.

**Q: How much storage do I need?**
A: It depends on your usage. Episodes are downloaded on-demand and cleaned up automatically based on your `CRON` schedule. Budget approximately 50-100MB per hour of audio.

**Q: Can I share my RSS feeds with others?**
A: Yes, but be aware that everyone using your instance will consume your API quota and storage.

**Q: Does this support video?**
A: No, CleanCast is audio-only to save storage and bandwidth.

---

## Roadmap

- [x] Playlist support
- [x] Channel support
- [x] SponsorBlock integration
- [x] Analytics dashboard
- [x] Smart playlists
- [x] Content filters
- [x] Backup/restore
- [x] Webhook notifications
- [ ] Web UI for management
- [ ] Discovery page via Apple Podcast API
- [ ] Multi-API key rotation
- [ ] iOS mobile app
- [ ] Android mobile app

See [open issues](https://github.com/ikoyhn/clean-cast/issues) for a full list of proposed features and known issues.

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

## Acknowledgments

- [lrstanley/go-ytdlp](https://github.com/lrstanley/go-ytdlp) - Go wrapper for yt-dlp
- [SponsorBlock](https://sponsor.ajay.app/) - Community-driven ad segment database
- [Google YouTube Data API](https://developers.google.com/youtube/v3) - YouTube metadata access
- [Echo Framework](https://echo.labstack.com/) - High-performance Go web framework
- [GORM](https://gorm.io/) - The fantastic ORM library for Golang

---

## Contact

Jared Lynch - jaredlynch13@gmail.com

Project Link: [https://github.com/ikoyhn/clean-cast](https://github.com/ikoyhn/clean-cast)

---

<div align="center">

Made with Go | Maintained with ❤️

</div>

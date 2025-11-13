# Changelog

All notable changes to CleanCast will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation including README, API reference, and deployment guides
- AWS SDK dependency for S3 backup functionality
- Documented all major dependencies in go.mod with comments

### Changed
- Enhanced README with complete feature list, architecture overview, and troubleshooting guide
- Improved environment variable documentation

## [2.0.0] - Recent Updates

### Added

#### Core Features
- **Analytics System**: Complete analytics tracking with dashboard
  - Episode play tracking with geographic distribution
  - Popular episodes reporting
  - Comprehensive analytics summary endpoint
  - Dashboard with aggregated metrics
- **Smart Playlists**: Create dynamic playlists based on custom filters
  - Filter by podcast, type, date range, duration
  - Generate RSS feeds from smart playlists
  - Full CRUD operations for smart playlist management
- **Content Filtering**: Advanced episode filtering capabilities
  - Keyword-based filters
  - Duration filters
  - Toggle filters on/off
  - Per-feed filter preferences
- **OPML Support**: Import and export podcast subscriptions
  - Import OPML files with multiple podcasts
  - Export current subscriptions to OPML
  - Compatible with standard podcast apps
- **Backup & Restore System**: Comprehensive backup solution
  - Database backup with JSON export
  - Optional audio file inclusion
  - S3 cloud storage integration
  - Scheduled automatic backups via cron
  - Restore from backup with database transaction safety
- **Webhook Notifications**: Event-driven notifications
  - Discord webhook support
  - Slack webhook support
  - Generic webhook support
  - Webhook delivery history tracking
  - Per-webhook event filtering
- **Transcript Support**: Access YouTube video transcripts
  - Fetch available transcripts
  - Multiple language support
  - Auto-generated and manual transcript detection
  - Search transcript content
- **Batch Operations**: Efficient bulk processing
  - Batch podcast refresh
  - Batch episode deletion
  - Batch podcast addition
  - Job status tracking
- **Search Functionality**: Full-text search capabilities
  - Episode search with filters
  - Podcast search
  - Support for complex query parameters

#### Technical Improvements
- **Prometheus Metrics**: Built-in monitoring with `/metrics` endpoint
- **Swagger Documentation**: Interactive API docs at `/docs`
- **Rate Limiting**: Configurable per-endpoint rate limits
- **Health Checks**: Kubernetes-ready `/health` and `/ready` endpoints
- **Request ID Middleware**: Request tracking across logs
- **Error Handling**: Structured error responses with details
- **Structured Logging**: Enhanced logging with zerolog
- **Multi-Format Audio**: Support for M4A, MP3, and Opus formats
- **Configurable Audio Quality**: Adjustable bitrate settings
- **Caching Layer**: In-memory caching for improved performance

#### User Preferences
- **Global Preferences**: User-wide default settings
- **Per-Feed Preferences**: Custom settings per podcast feed
- **Preference Management**: Full CRUD operations for preferences

### Changed
- **Removed Authentication**: Simplified deployment by removing host and token authentication
  - No more `TOKEN` environment variable required
  - No more `TRUSTED_HOSTS` configuration
  - All endpoints now use consistent rate limiting instead
- **Enhanced Docker Configuration**:
  - Multi-platform support (linux/amd64, linux/arm64)
  - Health check configuration
  - Resource limits and reservations
  - Security options (no-new-privileges)
  - Logging configuration
  - Custom networking
- **Improved Error Messages**: More descriptive error responses with context
- **Database Migrations**: Automated schema updates
- **Configuration Management**: Centralized config with environment variables

### Performance Improvements
- **On-Demand Downloads**: Episodes downloaded only when first accessed
- **Intelligent Caching**: Downloaded episodes cached for quick access
- **Automatic Cleanup**: Cron-based cleanup of old episodes
- **Database Optimization**: Indexed queries for faster searches
- **Parallel Processing**: Concurrent operations where applicable

### Developer Experience
- **Clean Architecture**: Well-organized codebase with clear separation
- **Comprehensive Tests**: Unit tests and benchmarks
- **Code Documentation**: Inline documentation and examples
- **API Standards**: RESTful API design with consistent patterns
- **Docker Support**: Production-ready containerization

### Bug Fixes
- Fixed GOOGLE_API_KEY handling in docker-compose.yml
- Removed unused string imports
- Fixed Docker image tagging for lowercase repository owners
- Improved error handling for file operations
- Enhanced validation for query parameters

### Security
- **Input Validation**: Comprehensive parameter validation
- **SQL Injection Protection**: GORM parameterized queries
- **Path Traversal Prevention**: Filename validation
- **Rate Limiting**: Protection against abuse
- **Security Headers**: Proper HTTP security headers
- **Read-Only Filesystem Support**: Optional Docker read-only mode

## Migration Guide

### From v1.x to v2.0

#### Environment Variables
The following environment variables have been **removed**:
- `TOKEN` - No longer used for authentication
- `TRUSTED_HOSTS` - No longer used for host validation

New **optional** environment variables:
- `AUDIO_FORMAT` - Set audio output format (default: m4a)
- `AUDIO_QUALITY` - Set audio bitrate (default: 192k)
- `BACKUP_CRON` - Schedule automatic backups
- `BACKUP_INCLUDE_AUDIO` - Include audio files in backups
- `BACKUP_S3_BUCKET` - S3 bucket for cloud backups
- `BACKUP_S3_REGION` - AWS region for S3
- `BACKUP_S3_ACCESS_KEY` - AWS access key
- `BACKUP_S3_SECRET_KEY` - AWS secret key

#### API Changes
No breaking changes to existing RSS feed endpoints:
- `GET /rss/:playlistId` - Still works as before
- `GET /channel/:channelId` - Still works as before
- `GET /media/:videoId` - Still works as before

New API endpoints are purely additive and don't affect existing integrations.

#### Docker Compose
Update your `docker-compose.yml` to include new environment variables:
```yaml
environment:
  - GOOGLE_API_KEY=${GOOGLE_API_KEY}
  - CRON=${CRON:-0 0 * * 0}
  - SPONSORBLOCK_CATEGORIES=${SPONSORBLOCK_CATEGORIES:-sponsor}
  - MIN_DURATION=${MIN_DURATION:-5m}
  - AUDIO_FORMAT=${AUDIO_FORMAT:-m4a}
  - AUDIO_QUALITY=${AUDIO_QUALITY:-192k}
```

Health checks are now included by default.

#### Database
Database migrations run automatically on startup. No manual intervention required.

#### Backups
If you want to enable automatic backups:
1. Set `BACKUP_CRON` environment variable (e.g., `0 3 * * *` for daily at 3 AM)
2. Optionally configure S3 variables for cloud storage
3. Set `BACKUP_INCLUDE_AUDIO=true` if you want audio files included

### Upgrading Steps

1. **Backup Your Data** (Important!)
   ```bash
   # Backup your config directory (includes database and audio)
   cp -r /path/to/config /path/to/config.backup
   ```

2. **Update Docker Image**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

3. **Remove Old Environment Variables**
   - Remove `TOKEN` from your `.env` file
   - Remove `TRUSTED_HOSTS` from your `.env` file

4. **Verify Health**
   ```bash
   curl http://localhost:8080/health
   ```

5. **Test Your RSS Feeds**
   - Check that existing RSS feeds still work
   - Verify audio playback

## Known Issues

None at this time. Please report issues on [GitHub](https://github.com/ikoyhn/clean-cast/issues).

## Planned Features

See the [Roadmap](README.md#roadmap) section in the README for upcoming features.

---

For more information, visit the [GitHub repository](https://github.com/ikoyhn/clean-cast).

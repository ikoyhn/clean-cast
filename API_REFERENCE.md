# CleanCast API Reference

Complete API documentation for CleanCast. For interactive documentation, visit `/docs` when running the application.

## Table of Contents

- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [RSS Feed Endpoints](#rss-feed-endpoints)
- [Media Streaming](#media-streaming)
- [Content Management](#content-management)
- [Smart Playlists](#smart-playlists)
- [User Preferences](#user-preferences)
- [Content Filters](#content-filters)
- [Search](#search)
- [Analytics](#analytics)
- [Backup & Restore](#backup--restore)
- [Batch Operations](#batch-operations)
- [Webhooks](#webhooks)
- [Transcripts](#transcripts)
- [System Endpoints](#system-endpoints)

---

## Authentication

**Current Version**: Authentication has been removed for simplified deployment.

All endpoints now use rate limiting instead of authentication. Future versions may reintroduce optional authentication.

---

## Rate Limiting

Rate limits are enforced per endpoint to prevent abuse:

| Endpoint Type | Rate Limit | Window |
|--------------|------------|--------|
| RSS Feeds | 10 req/min | 1 minute |
| Media Streaming | 50 req/min | 1 minute |
| Search | 20 req/min | 1 minute |
| Analytics | 20 req/min | 1 minute |
| Batch Operations | 5 req/min | 1 minute |
| Health Checks | 60 req/min | 1 minute |
| Other APIs | 10-20 req/min | 1 minute |

**Response Headers**:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Requests remaining in window
- `X-RateLimit-Reset`: Time when limit resets

**429 Response** when rate limit exceeded:
```json
{
  "error": "rate limit exceeded",
  "retry_after": 30
}
```

---

## Error Handling

All errors follow a consistent format:

### Error Response Format

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "additional context"
  }
}
```

### HTTP Status Codes

| Status | Description |
|--------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content (successful deletion) |
| 400 | Bad Request (invalid parameters) |
| 404 | Not Found |
| 429 | Too Many Requests (rate limited) |
| 500 | Internal Server Error |

### Common Error Codes

- `INVALID_PARAM` - Invalid parameter value
- `NOT_FOUND` - Resource not found
- `FILE_NOT_FOUND` - Audio file not found
- `BAD_REQUEST` - Invalid request format
- `INTERNAL_ERROR` - Server error

---

## RSS Feed Endpoints

### Get Playlist RSS Feed

Generate an RSS feed from a YouTube playlist.

**Endpoint**: `GET /rss/:playlistId`

**Parameters**:
- `playlistId` (path, required) - YouTube playlist ID

**Query Parameters**:
- `limit` (optional) - Maximum number of episodes (e.g., `?limit=50`)
- `date` (optional) - Filter episodes after date in MM-DD-YYYY format (e.g., `?date=01-01-2025`)
- `format` (optional) - Audio format: `m4a`, `mp3`, `opus` (default: from config)

**Note**: Cannot use both `limit` and `date` parameters together.

**Example Request**:
```bash
curl http://localhost:8080/rss/PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222
```

**Response**: RSS 2.0 XML feed

**Rate Limit**: 10 requests/minute

---

### Get Channel RSS Feed

Generate an RSS feed from all videos in a YouTube channel.

**Endpoint**: `GET /channel/:channelId`

**Parameters**:
- `channelId` (path, required) - YouTube channel ID

**Query Parameters**:
- `limit` (optional) - Maximum number of episodes
- `date` (optional) - Filter episodes after date in MM-DD-YYYY format
- `format` (optional) - Audio format: `m4a`, `mp3`, `opus`

**Example Request**:
```bash
curl "http://localhost:8080/channel/UCoj1ZgGoSBoonNZqMsVUfAA?date=01-01-2025"
```

**Response**: RSS 2.0 XML feed

**Rate Limit**: 10 requests/minute

**Warning**: Channel endpoints consume more YouTube API quota. Use the `date` parameter to reduce API usage.

---

## Media Streaming

### Stream Audio File

Stream or download an audio file for an episode.

**Endpoint**: `GET /media/:videoId`

**Parameters**:
- `videoId` (path, required) - YouTube video ID

**Query Parameters**:
- `format` (optional) - Audio format: `m4a`, `mp3`, `opus`

**Headers**:
- `Range` - Supports HTTP range requests for seeking

**Example Request**:
```bash
curl http://localhost:8080/media/dQw4w9WgXcQ
```

**Response**: Audio file stream with appropriate MIME type

**Behavior**:
1. If file exists and is recent, streams immediately
2. If file doesn't exist, downloads from YouTube first (may take time)
3. SponsorBlock segments are automatically removed
4. Subsequent requests are served from cache

**Rate Limit**: 50 requests/minute

---

## Content Management

### Import OPML

Import podcast subscriptions from an OPML file.

**Endpoint**: `POST /api/opml/import`

**Content-Type**: `multipart/form-data`

**Parameters**:
- `file` (form-data, required) - OPML file

**Example Request**:
```bash
curl -X POST http://localhost:8080/api/opml/import \
  -F "file=@subscriptions.opml"
```

**Response**:
```json
{
  "imported": 15,
  "failed": 2,
  "errors": [
    "Invalid feed URL: http://example.com/invalid"
  ]
}
```

**Rate Limit**: 5 requests/minute

---

### Export OPML

Export current podcast subscriptions to OPML format.

**Endpoint**: `GET /api/opml/export`

**Example Request**:
```bash
curl http://localhost:8080/api/opml/export > my-podcasts.opml
```

**Response**: OPML XML file

**Rate Limit**: 10 requests/minute

---

## Smart Playlists

### Create Smart Playlist

Create a dynamic playlist based on filters.

**Endpoint**: `POST /api/playlist/smart`

**Request Body**:
```json
{
  "name": "My Smart Playlist",
  "description": "Tech podcasts from this month",
  "rules": {
    "match_type": "all",
    "conditions": [
      {
        "field": "title",
        "operator": "contains",
        "value": "tech"
      },
      {
        "field": "duration",
        "operator": "greater_than",
        "value": "30m"
      },
      {
        "field": "published_date",
        "operator": "after",
        "value": "2025-01-01"
      }
    ]
  }
}
```

**Response**:
```json
{
  "id": "123",
  "name": "My Smart Playlist",
  "description": "Tech podcasts from this month",
  "rules": { ... },
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

**Rate Limit**: 10 requests/minute

---

### List Smart Playlists

Get all smart playlists.

**Endpoint**: `GET /api/playlist/smart`

**Response**:
```json
{
  "playlists": [
    {
      "id": "123",
      "name": "My Smart Playlist",
      "description": "Tech podcasts from this month",
      "episode_count": 45,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Smart Playlist

Get details of a specific smart playlist.

**Endpoint**: `GET /api/playlist/smart/:id`

**Response**:
```json
{
  "id": "123",
  "name": "My Smart Playlist",
  "description": "Tech podcasts from this month",
  "rules": { ... },
  "episodes": [ ... ],
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

**Rate Limit**: 20 requests/minute

---

### Update Smart Playlist

Update a smart playlist.

**Endpoint**: `PUT /api/playlist/smart/:id`

**Request Body**: Same as create

**Rate Limit**: 10 requests/minute

---

### Delete Smart Playlist

Delete a smart playlist.

**Endpoint**: `DELETE /api/playlist/smart/:id`

**Response**: 204 No Content

**Rate Limit**: 10 requests/minute

---

### Get Smart Playlist RSS Feed

Get RSS feed for a smart playlist.

**Endpoint**: `GET /rss/smart/:id`

**Response**: RSS 2.0 XML feed

**Rate Limit**: 10 requests/minute

---

## User Preferences

### Get User Preferences

Get global user preferences.

**Endpoint**: `GET /api/preferences`

**Response**:
```json
{
  "auto_download": false,
  "default_audio_format": "m4a",
  "default_audio_quality": "192k",
  "sponsorblock_categories": ["sponsor", "intro", "outro"]
}
```

**Rate Limit**: 20 requests/minute

---

### Update User Preferences

Update global user preferences.

**Endpoint**: `PUT /api/preferences`

**Request Body**:
```json
{
  "auto_download": true,
  "default_audio_format": "mp3",
  "default_audio_quality": "256k"
}
```

**Rate Limit**: 10 requests/minute

---

### Get Feed Preferences

Get preferences for a specific podcast feed.

**Endpoint**: `GET /api/preferences/feed/:feedId`

**Response**:
```json
{
  "feed_id": "PLbh0Jamvptwfp_qc439PLuyKJ-tWUt222",
  "audio_format": "opus",
  "audio_quality": "128k",
  "skip_intro": true,
  "skip_outro": true
}
```

**Rate Limit**: 20 requests/minute

---

### Update Feed Preferences

Update preferences for a specific podcast feed.

**Endpoint**: `PUT /api/preferences/feed/:feedId`

**Request Body**:
```json
{
  "audio_format": "opus",
  "audio_quality": "128k",
  "skip_intro": true
}
```

**Rate Limit**: 10 requests/minute

---

### Delete Feed Preferences

Remove custom preferences for a feed (reverts to global preferences).

**Endpoint**: `DELETE /api/preferences/feed/:feedId`

**Response**: 204 No Content

**Rate Limit**: 10 requests/minute

---

## Content Filters

### Create Filter

Create a content filter to exclude episodes.

**Endpoint**: `POST /api/filters`

**Request Body**:
```json
{
  "name": "Exclude shorts",
  "enabled": true,
  "filter_type": "duration",
  "pattern": "<5m",
  "apply_to": ["all"]
}
```

**Filter Types**:
- `keyword` - Filter by title/description keywords
- `duration` - Filter by episode duration
- `regex` - Advanced regex pattern matching

**Rate Limit**: 10 requests/minute

---

### List Filters

Get all content filters.

**Endpoint**: `GET /api/filters`

**Response**:
```json
{
  "filters": [
    {
      "id": "1",
      "name": "Exclude shorts",
      "enabled": true,
      "filter_type": "duration",
      "pattern": "<5m",
      "match_count": 127,
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Filter

Get a specific filter.

**Endpoint**: `GET /api/filters/:id`

**Rate Limit**: 20 requests/minute

---

### Update Filter

Update a content filter.

**Endpoint**: `PUT /api/filters/:id`

**Request Body**: Same as create

**Rate Limit**: 10 requests/minute

---

### Delete Filter

Delete a content filter.

**Endpoint**: `DELETE /api/filters/:id`

**Response**: 204 No Content

**Rate Limit**: 10 requests/minute

---

### Toggle Filter

Enable or disable a filter.

**Endpoint**: `PATCH /api/filters/:id/toggle`

**Request Body**:
```json
{
  "enabled": false
}
```

**Rate Limit**: 10 requests/minute

---

## Search

### Search Episodes

Search for episodes across all podcasts.

**Endpoint**: `GET /search/episodes`

**Query Parameters**:
- `query` (required) - Search query string
- `podcast_id` (optional) - Limit to specific podcast
- `type` (optional) - `CHANNEL` or `PLAYLIST`
- `start_date` (optional) - YYYY-MM-DD format
- `end_date` (optional) - YYYY-MM-DD format
- `min_duration` (optional) - Minimum duration (e.g., `30m`)
- `max_duration` (optional) - Maximum duration (e.g., `2h`)
- `limit` (optional) - Results limit (default: 50)
- `offset` (optional) - Results offset for pagination

**Example Request**:
```bash
curl "http://localhost:8080/search/episodes?query=interview&min_duration=30m&limit=20"
```

**Response**:
```json
{
  "results": [
    {
      "video_id": "abc123",
      "title": "Tech Interview with John Doe",
      "podcast_title": "Tech Talks",
      "duration": 3600,
      "published_date": "2025-01-15T10:00:00Z",
      "thumbnail_url": "https://..."
    }
  ],
  "total": 156,
  "limit": 20,
  "offset": 0
}
```

**Rate Limit**: 20 requests/minute

---

### Search Podcasts

Search for podcasts.

**Endpoint**: `GET /search/podcasts`

**Query Parameters**:
- `query` (required) - Search query string
- `limit` (optional) - Results limit (default: 50)
- `offset` (optional) - Results offset

**Example Request**:
```bash
curl "http://localhost:8080/search/podcasts?query=technology"
```

**Response**:
```json
{
  "results": [
    {
      "id": "PLabc123",
      "title": "Tech Weekly",
      "type": "PLAYLIST",
      "episode_count": 250,
      "channel_name": "Tech Channel"
    }
  ],
  "total": 45,
  "limit": 50,
  "offset": 0
}
```

**Rate Limit**: 20 requests/minute

---

## Analytics

### Get Popular Episodes

Get most played episodes.

**Endpoint**: `GET /api/analytics/popular`

**Query Parameters**:
- `limit` (optional) - Number of results (default: 20)
- `days` (optional) - Time window in days (default: 30)

**Example Request**:
```bash
curl "http://localhost:8080/api/analytics/popular?limit=10&days=7"
```

**Response**:
```json
{
  "episodes": [
    {
      "video_id": "abc123",
      "title": "Most Popular Episode",
      "play_count": 1543,
      "unique_listeners": 892,
      "avg_completion": 0.85
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Episode Analytics

Get analytics for a specific episode.

**Endpoint**: `GET /api/analytics/episode/:videoId`

**Response**:
```json
{
  "video_id": "abc123",
  "title": "Episode Title",
  "total_plays": 1543,
  "unique_listeners": 892,
  "avg_completion": 0.85,
  "play_duration_total": 4629000,
  "geographic_distribution": {
    "US": 650,
    "GB": 230,
    "CA": 180
  },
  "plays_by_day": [
    {
      "date": "2025-01-15",
      "plays": 89
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Analytics Summary

Get overall analytics summary.

**Endpoint**: `GET /api/analytics/summary`

**Query Parameters**:
- `days` (optional) - Time window in days (default: 30)

**Response**:
```json
{
  "total_plays": 45892,
  "unique_listeners": 12456,
  "total_podcasts": 45,
  "total_episodes": 2341,
  "avg_episode_duration": 3600,
  "total_listening_time": 165312000,
  "top_podcast": {
    "id": "PLabc123",
    "title": "Tech Weekly",
    "play_count": 5432
  }
}
```

**Rate Limit**: 20 requests/minute

---

### Get Geographic Distribution

Get listener geographic distribution.

**Endpoint**: `GET /api/analytics/geographic`

**Query Parameters**:
- `days` (optional) - Time window in days (default: 30)

**Response**:
```json
{
  "total_plays": 45892,
  "countries": [
    {
      "code": "US",
      "name": "United States",
      "plays": 18356,
      "percentage": 40.0
    },
    {
      "code": "GB",
      "name": "United Kingdom",
      "plays": 9178,
      "percentage": 20.0
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Dashboard Data

Get comprehensive analytics for dashboard display.

**Endpoint**: `GET /api/analytics/dashboard`

**Query Parameters**:
- `days` (optional) - Time window in days (default: 7)

**Response**: Combines summary, popular episodes, geographic data, and trends.

**Rate Limit**: 20 requests/minute

---

## Backup & Restore

### Create Backup

Create a backup of database and optionally audio files.

**Endpoint**: `POST /api/backup/create`

**Request Body**:
```json
{
  "include_audio": false,
  "description": "Manual backup before upgrade"
}
```

**Response**:
```json
{
  "backup_id": "backup_20250115_143000",
  "size": 52428800,
  "file_count": 0,
  "created_at": "2025-01-15T14:30:00Z",
  "description": "Manual backup before upgrade"
}
```

**Rate Limit**: 5 requests/minute

---

### List Backups

Get list of all backups.

**Endpoint**: `GET /api/backup/list`

**Response**:
```json
{
  "backups": [
    {
      "id": "backup_20250115_143000",
      "size": 52428800,
      "file_count": 0,
      "created_at": "2025-01-15T14:30:00Z",
      "description": "Manual backup before upgrade"
    }
  ]
}
```

**Rate Limit**: 10 requests/minute

---

### Restore Backup

Restore from a backup.

**Endpoint**: `POST /api/backup/restore/:backupId`

**Request Body**:
```json
{
  "include_audio": false
}
```

**Response**:
```json
{
  "success": true,
  "podcasts_restored": 45,
  "episodes_restored": 2341,
  "audio_files_restored": 0
}
```

**Warning**: This operation will overwrite existing data. Create a backup first!

**Rate Limit**: 5 requests/minute

---

### Delete Backup

Delete a backup file.

**Endpoint**: `DELETE /api/backup/:backupId`

**Response**: 204 No Content

**Rate Limit**: 10 requests/minute

---

### Upload Backup to S3

Upload a backup to S3 cloud storage.

**Endpoint**: `POST /api/backup/upload/:backupId`

**Request Body**:
```json
{
  "bucket": "my-backups",
  "region": "us-east-1",
  "access_key": "AKIA...",
  "secret_key": "secret..."
}
```

**Response**:
```json
{
  "success": true,
  "s3_key": "backup_20250115_143000.zip"
}
```

**Rate Limit**: 5 requests/minute

---

### Download Backup from S3

Download a backup from S3.

**Endpoint**: `POST /api/backup/download`

**Request Body**:
```json
{
  "backup_id": "backup_20250115_143000",
  "bucket": "my-backups",
  "region": "us-east-1",
  "access_key": "AKIA...",
  "secret_key": "secret..."
}
```

**Rate Limit**: 5 requests/minute

---

## Batch Operations

### Batch Refresh Podcasts

Refresh metadata for multiple podcasts.

**Endpoint**: `POST /api/batch/refresh`

**Request Body**:
```json
{
  "podcast_ids": ["PLabc123", "PLdef456", "PLghi789"]
}
```

**Response**:
```json
{
  "job_id": "job_abc123",
  "status": "processing",
  "total": 3,
  "completed": 0
}
```

**Rate Limit**: 5 requests/minute

---

### Batch Delete Episodes

Delete multiple episodes at once.

**Endpoint**: `POST /api/batch/episodes/delete`

**Request Body**:
```json
{
  "video_ids": ["abc123", "def456", "ghi789"]
}
```

**Response**:
```json
{
  "job_id": "job_def456",
  "status": "processing",
  "total": 3,
  "completed": 0
}
```

**Rate Limit**: 5 requests/minute

---

### Batch Add Podcasts

Add multiple podcasts at once.

**Endpoint**: `POST /api/batch/podcasts/add`

**Request Body**:
```json
{
  "podcasts": [
    {
      "type": "playlist",
      "id": "PLabc123"
    },
    {
      "type": "channel",
      "id": "UCdef456"
    }
  ]
}
```

**Response**:
```json
{
  "job_id": "job_ghi789",
  "status": "processing",
  "total": 2,
  "completed": 0
}
```

**Rate Limit**: 5 requests/minute

---

### Get Batch Job Status

Check the status of a batch operation.

**Endpoint**: `GET /api/batch/status/:jobId`

**Response**:
```json
{
  "job_id": "job_abc123",
  "status": "completed",
  "total": 3,
  "completed": 3,
  "failed": 0,
  "errors": [],
  "started_at": "2025-01-15T14:30:00Z",
  "completed_at": "2025-01-15T14:32:15Z"
}
```

**Status Values**:
- `pending` - Job queued
- `processing` - Job in progress
- `completed` - Job finished successfully
- `failed` - Job failed
- `partial` - Job completed with some failures

**Rate Limit**: 20 requests/minute

---

## Webhooks

### Create Webhook

Create a webhook configuration.

**Endpoint**: `POST /api/webhooks`

**Request Body**:
```json
{
  "name": "Discord Notifications",
  "url": "https://discord.com/api/webhooks/...",
  "type": "discord",
  "events": "episode.new,podcast.updated",
  "enabled": true
}
```

**Webhook Types**:
- `discord` - Discord webhook
- `slack` - Slack webhook
- `generic` - Generic JSON webhook

**Events** (comma-separated):
- `episode.new` - New episode available
- `episode.downloaded` - Episode downloaded
- `podcast.updated` - Podcast metadata updated
- `backup.completed` - Backup completed
- `error` - Error occurred

**Response**:
```json
{
  "id": "1",
  "name": "Discord Notifications",
  "url": "https://discord.com/api/webhooks/...",
  "type": "discord",
  "events": "episode.new,podcast.updated",
  "enabled": true,
  "created_at": "2025-01-15T14:30:00Z"
}
```

**Rate Limit**: 10 requests/minute

---

### List Webhooks

Get all webhook configurations.

**Endpoint**: `GET /api/webhooks`

**Rate Limit**: 20 requests/minute

---

### Get Webhook

Get a specific webhook configuration.

**Endpoint**: `GET /api/webhooks/:id`

**Rate Limit**: 20 requests/minute

---

### Update Webhook

Update a webhook configuration.

**Endpoint**: `PUT /api/webhooks/:id`

**Request Body**: Same as create

**Rate Limit**: 10 requests/minute

---

### Delete Webhook

Delete a webhook configuration.

**Endpoint**: `DELETE /api/webhooks/:id`

**Response**: 204 No Content

**Rate Limit**: 10 requests/minute

---

### Get Webhook Deliveries

Get delivery history for a webhook.

**Endpoint**: `GET /api/webhooks/:id/deliveries`

**Query Parameters**:
- `limit` (optional) - Number of results (default: 50, max: 100)

**Response**:
```json
{
  "deliveries": [
    {
      "id": "123",
      "webhook_id": "1",
      "event": "episode.new",
      "status": "success",
      "status_code": 200,
      "delivered_at": "2025-01-15T14:30:00Z",
      "response_time_ms": 145
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

## Transcripts

### Get Transcript

Get transcript for a video.

**Endpoint**: `GET /api/transcript/:videoId`

**Query Parameters**:
- `lang` (optional) - Language code (e.g., `en`, `es`)

**Response**:
```json
{
  "video_id": "abc123",
  "language": "en",
  "language_name": "English",
  "is_auto_generated": false,
  "segments": [
    {
      "start": 0.0,
      "duration": 2.5,
      "text": "Welcome to this episode"
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get All Transcripts

Get all available transcripts for a video.

**Endpoint**: `GET /api/transcript/:videoId/all`

**Response**:
```json
{
  "video_id": "abc123",
  "transcripts": [
    {
      "language": "en",
      "language_name": "English",
      "is_auto_generated": false
    },
    {
      "language": "es",
      "language_name": "Spanish",
      "is_auto_generated": true
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Get Available Languages

Get list of available transcript languages.

**Endpoint**: `GET /api/transcript/:videoId/languages`

**Response**:
```json
{
  "video_id": "abc123",
  "languages": [
    {
      "code": "en",
      "name": "English",
      "auto_generated": false
    }
  ]
}
```

**Rate Limit**: 20 requests/minute

---

### Fetch Transcript

Explicitly fetch and cache a transcript from YouTube.

**Endpoint**: `POST /api/transcript/:videoId/fetch`

**Request Body**:
```json
{
  "language": "en"
}
```

**Rate Limit**: 10 requests/minute

---

## System Endpoints

### Health Check

Check if the service is healthy.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": 3600,
  "database": "connected"
}
```

**Rate Limit**: 60 requests/minute

---

### Readiness Check

Check if the service is ready to accept requests.

**Endpoint**: `GET /ready`

**Response**:
```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "storage": "ok",
    "youtube_api": "ok"
  }
}
```

**Rate Limit**: 60 requests/minute

---

### Root Endpoint

Basic service information.

**Endpoint**: `GET /`

**Response**: `Hello, World!`

**Rate Limit**: None

---

### Prometheus Metrics

Prometheus-formatted metrics for monitoring.

**Endpoint**: `GET /metrics`

**Response**: Prometheus text format

**Metrics Include**:
- HTTP request duration
- Request count by endpoint
- Error rates
- Active connections
- Episode download statistics
- Cache hit rates

**Rate Limit**: None

---

### Swagger Documentation

Interactive API documentation.

**Endpoint**: `GET /docs/*`

**Description**: Swagger UI for exploring and testing the API

**Rate Limit**: None

---

## Request ID Tracking

All requests include a `X-Request-ID` header in responses for tracking and debugging:

```bash
curl -I http://localhost:8080/health
X-Request-ID: abc123def456
```

Use this ID when reporting issues or checking logs.

---

## CORS Support

Cross-Origin Resource Sharing (CORS) is enabled by default for all origins. Configure as needed in production.

---

## Content Types

### Supported Audio Formats

| Format | MIME Type | Extension | Typical Bitrate |
|--------|-----------|-----------|----------------|
| M4A | `audio/mp4` | `.m4a` | 192k |
| MP3 | `audio/mpeg` | `.mp3` | 192k |
| Opus | `audio/opus` | `.opus` | 128k |

### RSS Feed Format

- Follows RSS 2.0 specification
- Compatible with iTunes podcast extensions
- Includes full episode metadata and artwork

---

## Best Practices

1. **Use the `date` parameter** on channel endpoints to reduce API quota usage
2. **Cache RSS feeds** on the client side (most podcast apps do this automatically)
3. **Implement retry logic** with exponential backoff for failed requests
4. **Monitor rate limits** using response headers
5. **Use webhooks** for real-time notifications instead of polling
6. **Batch operations** when performing bulk updates
7. **Set appropriate audio format** based on your use case:
   - M4A: Best quality, larger files
   - MP3: Universal compatibility
   - Opus: Best compression, smaller files

---

## Support

For issues, questions, or feature requests:
- GitHub Issues: https://github.com/ikoyhn/clean-cast/issues
- Discussions: https://github.com/ikoyhn/clean-cast/discussions

---

**Version**: 2.0.0
**Last Updated**: 2025-01-15

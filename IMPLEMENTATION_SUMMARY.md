# Backup/Restore and Multi-Format Support Implementation Summary

This document outlines the implementation of backup/restore functionality and multi-format audio support for the clean-cast project.

## Features Implemented

### 1. Backup & Restore Service

**File:** `/home/user/clean-cast/internal/services/backup/backup.go`

#### Features:
- **Export Database to JSON**: Full database export including podcasts, episodes, and playback history
- **Export Audio Files List**: Tracks all audio files in the system
- **Create Backups**: Creates timestamped ZIP archives containing database and optionally audio files
- **Restore from Backup**: Restores database and audio files from backup archives
- **List Backups**: View all available backups with metadata
- **Delete Backups**: Remove old backups
- **S3/Cloud Storage Support**: Upload and download backups to/from AWS S3
- **Scheduled Backups**: Integration with cron for automated backups

#### Key Functions:
```go
- ExportDatabase() (*BackupData, error)
- GetAudioFilesList() ([]string, error)
- CreateBackup(includeAudio bool, description string) (string, error)
- ListBackups() ([]BackupMetadata, error)
- RestoreFromBackup(backupID string, includeAudio bool) error
- DeleteBackup(backupID string) error
- UploadToS3(...) error
- DownloadFromS3(...) error
- ScheduledBackup(includeAudio bool) error
```

### 2. Backup API Endpoints

**File:** `/home/user/clean-cast/internal/api/backup.go`

#### Endpoints:

- `POST /api/backup/create` - Create a new backup
  ```json
  {
    "include_audio": true,
    "description": "Manual backup before update"
  }
  ```

- `GET /api/backup/list` - List all available backups
  ```json
  {
    "backups": [
      {
        "id": "backup_20251113_143025",
        "created_at": "2025-11-13T14:30:25Z",
        "size": 1048576,
        "file_count": 150,
        "description": "Scheduled backup"
      }
    ],
    "count": 1
  }
  ```

- `POST /api/backup/restore` - Restore from a backup
  ```json
  {
    "backup_id": "backup_20251113_143025",
    "include_audio": true
  }
  ```

- `GET /api/backup/download/:id` - Download a backup file

- `DELETE /api/backup/:id` - Delete a backup

- `POST /api/backup/upload-s3` - Upload backup to S3
  ```json
  {
    "backup_id": "backup_20251113_143025",
    "bucket": "my-backups",
    "region": "us-east-1",
    "access_key": "...",
    "secret_key": "..."
  }
  ```

- `POST /api/backup/download-s3` - Download backup from S3

### 3. Multi-Format Audio Support

**File:** `/home/user/clean-cast/internal/services/downloader/ytdlpService.go`

#### Supported Formats:
- **M4A** (audio/mp4) - Default, 192k quality
- **MP3** (audio/mpeg) - 192k quality
- **Opus** (audio/opus) - 128k quality

#### New Functions:
```go
type AudioFormat struct {
    Format      string // m4a, mp3, opus
    Quality     string // 128k, 192k, 320k
    Extension   string // .m4a, .mp3, .opus
    MimeType    string // MIME type
    FormatSort  string // yt-dlp format sort string
}

func GetAudioFormat(format string) AudioFormat
func GetYoutubeVideoWithFormat(youtubeVideoId string, audioFormat AudioFormat) (string, <-chan struct{})
```

### 4. Format-Aware RSS Feeds

**Files:**
- `/home/user/clean-cast/internal/services/rss/rssService.go`
- `/home/user/clean-cast/internal/services/playlist/playlistService.go`

#### Changes:
- RSS feed generators now accept audio format parameter
- Media URLs include format in the enclosure
- Correct MIME types set based on format
- Cache includes format in cache key

### 5. Configuration Updates

**File:** `/home/user/clean-cast/internal/config/config.go`

#### New Environment Variables:

```bash
# Audio format settings
AUDIO_FORMAT=m4a          # Default: m4a. Options: m4a, mp3, opus
AUDIO_QUALITY=192k        # Default: 192k. Options: 128k, 192k, 320k

# Backup settings
BACKUP_CRON="0 2 * * *"   # Cron schedule for automated backups
BACKUP_INCLUDE_AUDIO=true # Whether to include audio files in backups

# S3 backup settings (optional)
BACKUP_S3_BUCKET=my-backups
BACKUP_S3_REGION=us-east-1
BACKUP_S3_ACCESS_KEY=your-access-key
BACKUP_S3_SECRET_KEY=your-secret-key
```

### 6. Media Endpoint Updates

**File:** `/home/user/clean-cast/internal/app/controller.go`

#### Query Parameters:
```
GET /media/:videoId?format=mp3&quality=320k
```

- `format` - Audio format (m4a, mp3, opus). Falls back to AUDIO_FORMAT env var
- `quality` - Audio quality (128k, 192k, 320k). Falls back to AUDIO_QUALITY env var

#### RSS Feed Query Parameters:
```
GET /channel/:channelId?format=mp3
GET /rss/:playlistId?format=mp3
```

The RSS feed will include media URLs with the specified format.

## Testing

### Test Multi-Format Support

1. **Request MP3 format:**
   ```bash
   curl http://localhost:8080/media/VIDEO_ID?format=mp3&token=YOUR_TOKEN
   ```

2. **Request Opus format:**
   ```bash
   curl http://localhost:8080/media/VIDEO_ID?format=opus&token=YOUR_TOKEN
   ```

3. **RSS feed with MP3:**
   ```bash
   curl http://localhost:8080/channel/CHANNEL_ID?format=mp3&token=YOUR_TOKEN
   ```

### Test Backup/Restore

1. **Create a backup:**
   ```bash
   curl -X POST http://localhost:8080/api/backup/create \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"include_audio": false, "description": "Test backup"}'
   ```

2. **List backups:**
   ```bash
   curl http://localhost:8080/api/backup/list \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

3. **Download a backup:**
   ```bash
   curl http://localhost:8080/api/backup/download/backup_ID \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -o backup.zip
   ```

4. **Restore from backup:**
   ```bash
   curl -X POST http://localhost:8080/api/backup/restore \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"backup_id": "backup_ID", "include_audio": false}'
   ```

## Dependencies Required

The following Go module needs to be added to `go.mod`:

```bash
go get github.com/aws/aws-sdk-go
```

This provides the AWS SDK for S3 backup functionality. If S3 is not needed, the S3-related functions can be removed from `backup.go`.

## Cron Job Integration

The backup cron job is automatically registered in `setupCron()` function in `controller.go`:

```go
if config.Config.BackupCron != "" {
    c.AddFunc(config.Config.BackupCron, func() {
        logger.Logger.Info().Msg("Running scheduled backup")
        if err := backup.ScheduledBackup(config.Config.BackupIncludeAudio); err != nil {
            logger.Logger.Error().Err(err).Msg("Scheduled backup failed")
        }
    })
}
```

Example cron schedules:
- `"0 2 * * *"` - Daily at 2 AM
- `"0 */6 * * *"` - Every 6 hours
- `"0 0 * * 0"` - Weekly on Sunday at midnight

## File Structure

```
internal/
├── api/
│   └── backup.go               # Backup API handlers
├── services/
│   ├── backup/
│   │   └── backup.go          # Backup/restore service
│   ├── downloader/
│   │   └── ytdlpService.go    # Multi-format download support
│   ├── rss/
│   │   └── rssService.go      # Format-aware RSS generation
│   └── playlist/
│       └── playlistService.go # Format-aware playlist RSS
├── config/
│   └── config.go              # Configuration with new env vars
└── app/
    └── controller.go          # Route registration and handlers
```

## Backup File Structure

Backups are stored in `/config/backups/` with the following structure:

```
/config/backups/
├── backup_20251113_143025.zip      # Backup archive
├── backup_20251113_143025.json     # Backup metadata
└── ...
```

Backup ZIP contents:
```
backup_20251113_143025.zip
├── database.json                    # Full database export
└── audio/                          # Audio files (if included)
    ├── VIDEO_ID1.m4a
    ├── VIDEO_ID2.mp3
    └── ...
```

## Known Limitations

1. **Format Conversion**: When switching formats, existing files are not automatically converted. Users need to re-download or the system will download on first request.

2. **Storage**: Different formats will result in separate downloads. If using multiple formats, disk space usage will increase.

3. **S3 Dependencies**: S3 functionality requires AWS SDK. If not needed, consider removing or making it optional with build tags.

4. **Cache Invalidation**: When changing formats via query parameters, the RSS feed cache needs to account for format in the cache key (already implemented).

## Future Enhancements

1. **Format Conversion**: Add ability to convert existing audio files between formats
2. **Per-Feed Format Preferences**: Store format preferences per feed in database
3. **Automatic Cleanup**: Remove old backups automatically based on retention policy
4. **Incremental Backups**: Support incremental backups to reduce backup size
5. **Backup Encryption**: Add encryption support for sensitive backups
6. **Other Cloud Providers**: Support Google Cloud Storage, Azure Blob Storage, etc.

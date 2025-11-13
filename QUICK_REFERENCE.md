# Quick Reference Guide - Backup/Restore and Multi-Format Support

## Files Created

### New Files
1. `/home/user/clean-cast/internal/services/backup/backup.go` - Backup/restore service
2. `/home/user/clean-cast/internal/api/backup.go` - Backup API handlers
3. `/home/user/clean-cast/IMPLEMENTATION_SUMMARY.md` - Detailed implementation documentation
4. `/home/user/clean-cast/docker-compose.example.yml` - Docker compose example with new env vars
5. `/home/user/clean-cast/CONTROLLER_RSS_PATCH.md` - Manual patches needed for controller.go
6. `/home/user/clean-cast/QUICK_REFERENCE.md` - This file

## Files Modified

### Core Services
1. `/home/user/clean-cast/internal/services/downloader/ytdlpService.go`
   - Added `AudioFormat` struct
   - Added `GetAudioFormat()` function
   - Added `GetYoutubeVideoWithFormat()` function
   - Modified download logic to support multiple formats

2. `/home/user/clean-cast/internal/services/rss/rssService.go`
   - Updated `GenerateRssFeed()` to accept audioFormat parameter
   - Updated `BuildChannelRssFeed()` to accept audioFormat parameter
   - Added `getEnclosureType()` helper function
   - Format-aware media URLs in RSS feeds

3. `/home/user/clean-cast/internal/services/playlist/playlistService.go`
   - Updated `BuildPlaylistRssFeed()` to accept audioFormat parameter

### Configuration
4. `/home/user/clean-cast/internal/config/config.go`
   - Added `AudioFormat` field
   - Added `AudioQuality` field
   - Added `BackupCron` field
   - Added `BackupIncludeAudio` field
   - Added environment variable loading for new config

### Application
5. `/home/user/clean-cast/internal/app/controller.go`
   - Added backup route registration
   - Updated media endpoint to support format query parameter
   - Updated cron setup to include backup scheduling
   - **NOTE:** RSS endpoints need manual patching (see CONTROLLER_RSS_PATCH.md)

## Environment Variables Reference

```bash
# Audio Format (NEW)
AUDIO_FORMAT=m4a          # Default format: m4a, mp3, opus
AUDIO_QUALITY=192k        # Default quality: 128k, 192k, 320k

# Backup Configuration (NEW)
BACKUP_CRON="0 2 * * *"   # Cron schedule for automated backups
BACKUP_INCLUDE_AUDIO=true # Include audio files in backups

# S3 Backup (NEW, Optional)
BACKUP_S3_BUCKET=my-backups
BACKUP_S3_REGION=us-east-1
BACKUP_S3_ACCESS_KEY=key
BACKUP_S3_SECRET_KEY=secret

# Existing Variables
GOOGLE_API_KEY=required
TOKEN=optional
SPONSORBLOCK_CATEGORIES=sponsor,intro,outro
CRON="0 0 * * *"
MIN_DURATION=2m
COOKIES_FILE=cookies.txt
```

## API Endpoints Reference

### Backup Endpoints (NEW)

```bash
# Create backup
POST /api/backup/create
Body: {"include_audio": true, "description": "Manual backup"}

# List backups
GET /api/backup/list

# Restore backup
POST /api/backup/restore
Body: {"backup_id": "backup_20251113_143025", "include_audio": true}

# Download backup
GET /api/backup/download/:id

# Delete backup
DELETE /api/backup/:id

# Upload to S3
POST /api/backup/upload-s3
Body: {"backup_id": "backup_ID", "bucket": "...", "region": "...", "access_key": "...", "secret_key": "..."}

# Download from S3
POST /api/backup/download-s3
Body: {"backup_id": "backup_ID", "bucket": "...", "region": "...", "access_key": "...", "secret_key": "..."}
```

### Format Query Parameters (NEW)

```bash
# Media endpoint with format
GET /media/:videoId?format=mp3&quality=320k

# Channel RSS with format
GET /channel/:channelId?format=mp3

# Playlist RSS with format
GET /rss/:playlistId?format=opus
```

## Quick Start Testing

### 1. Test Multi-Format Downloads

```bash
# Download as MP3
curl "http://localhost:8080/media/VIDEO_ID?format=mp3&token=TOKEN" -o test.mp3

# Download as Opus
curl "http://localhost:8080/media/VIDEO_ID?format=opus&token=TOKEN" -o test.opus

# Download as M4A (default)
curl "http://localhost:8080/media/VIDEO_ID?token=TOKEN" -o test.m4a
```

### 2. Test RSS Feeds with Format

```bash
# Get RSS feed with MP3 format
curl "http://localhost:8080/channel/CHANNEL_ID?format=mp3&token=TOKEN"

# Get RSS feed with default format
curl "http://localhost:8080/channel/CHANNEL_ID?token=TOKEN"
```

### 3. Test Backup/Restore

```bash
# Create backup (database only)
curl -X POST http://localhost:8080/api/backup/create \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"include_audio": false, "description": "Test backup"}'

# List backups
curl http://localhost:8080/api/backup/list \
  -H "Authorization: Bearer TOKEN"

# Download backup
curl http://localhost:8080/api/backup/download/backup_ID \
  -H "Authorization: Bearer TOKEN" \
  -o backup.zip
```

## Deployment Checklist

- [ ] Add AWS SDK dependency: `go get github.com/aws/aws-sdk-go` (if using S3)
- [ ] Apply controller.go patches from CONTROLLER_RSS_PATCH.md
- [ ] Update docker-compose.yml with new environment variables
- [ ] Set AUDIO_FORMAT and AUDIO_QUALITY env vars
- [ ] Configure BACKUP_CRON for automated backups (optional)
- [ ] Configure S3 credentials if using cloud backups (optional)
- [ ] Test format conversion with FFmpeg
- [ ] Verify backup/restore functionality
- [ ] Test RSS feeds with different formats
- [ ] Monitor disk space (multiple formats increase storage)

## Format Support Matrix

| Format | Extension | MIME Type     | Default Quality | Notes                    |
|--------|-----------|---------------|-----------------|--------------------------|
| M4A    | .m4a      | audio/mp4     | 192k            | Default, best quality    |
| MP3    | .mp3      | audio/mpeg    | 192k            | Universal compatibility  |
| Opus   | .opus     | audio/opus    | 128k            | Best compression         |

## Backup Directory Structure

```
/config/
├── backups/
│   ├── backup_20251113_143025.zip     # Backup archive
│   ├── backup_20251113_143025.json    # Backup metadata
│   └── ...
├── audio/
│   ├── VIDEO_ID1.m4a
│   ├── VIDEO_ID2.mp3
│   └── ...
└── sqlite.db
```

## Cron Schedule Examples

```bash
# Every day at 2 AM
BACKUP_CRON="0 2 * * *"

# Every 6 hours
BACKUP_CRON="0 */6 * * *"

# Every Sunday at midnight
BACKUP_CRON="0 0 * * 0"

# Twice a day (8 AM and 8 PM)
BACKUP_CRON="0 8,20 * * *"

# Every Monday and Thursday at 3 AM
BACKUP_CRON="0 3 * * 1,4"
```

## Troubleshooting

### Issue: Format conversion fails
**Solution:** Ensure FFmpeg is installed and accessible at `/usr/bin/ffmpeg`

### Issue: Backup fails with disk space error
**Solution:** Check available disk space. Consider setting `BACKUP_INCLUDE_AUDIO=false` or using S3

### Issue: S3 upload fails
**Solution:** Verify S3 credentials and bucket permissions. Check network connectivity

### Issue: Different format not downloading
**Solution:** Check yt-dlp logs. Some videos may not support all formats

### Issue: RSS feed shows wrong MIME type
**Solution:** Ensure controller.go patches are applied correctly

## Performance Considerations

1. **Storage**: Each format creates a separate file. A video with 3 formats uses 3x storage
2. **Download Time**: First request for new format will download from YouTube
3. **Backup Size**: Including audio files significantly increases backup size
4. **Cache**: RSS feeds are cached per format, reducing database load

## Security Notes

1. Backup API endpoints require authentication (TOKEN)
2. S3 credentials can be set via environment variables or API requests
3. Backup files contain full database, protect access appropriately
4. Consider encrypting backups for sensitive data

## Next Steps

1. Review IMPLEMENTATION_SUMMARY.md for detailed documentation
2. Apply patches from CONTROLLER_RSS_PATCH.md
3. Test all functionality thoroughly
4. Update your deployment configuration
5. Monitor logs for any issues
6. Consider setting up automated backup monitoring

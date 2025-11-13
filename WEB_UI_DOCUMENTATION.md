# Clean Cast Web UI Admin Panel Documentation

## Overview

A comprehensive, modern web-based admin panel for managing your Clean Cast podcast system. Built with vanilla JavaScript, Tailwind CSS, Chart.js, and a custom API client.

## Features

### 1. Dashboard (`/admin`)
- **Real-time Statistics**: Total podcasts, episodes, plays, and storage usage
- **Visual Charts**:
  - Plays over time (last 7 days)
  - Most popular episodes
- **Recent Activity Feed**: Latest episode plays and interactions
- **Quick Actions**:
  - Add podcast
  - Refresh all podcasts
  - Create backup
  - View analytics
  - View logs
- **System Health Indicators**: API status, database connection, cache status
- **Active Jobs Monitor**: Track batch operations in real-time

### 2. Podcast Management (`/admin/podcasts`)
- **Podcast List**: View all subscribed podcasts with pagination
- **Search Functionality**: Search by name, description, or artist
- **Bulk Operations**:
  - Add multiple podcasts at once
  - Refresh selected podcasts
  - Bulk selection with checkboxes
- **Individual Actions**:
  - View episodes
  - Refresh metadata
  - Copy RSS feed URL
  - View details
- **Add Podcast Modal**:
  - Support for YouTube channels and playlists
  - Auto-detection of podcast type
  - URL or ID input

### 3. Episode Management (`/admin/episodes`)
- **Episode Browser**: Paginated list of all episodes
- **Advanced Filtering**:
  - By podcast
  - By type (channel/playlist)
  - Date range
  - Duration range
  - Search by title/description
- **Bulk Operations**:
  - Delete multiple episodes
  - Batch selection
- **Episode Actions**:
  - Play episode
  - View analytics
  - View transcript
  - Delete episode
- **Visual Information**: Thumbnails, duration, play counts, publish dates

### 4. Settings (`/admin/settings`)

#### Webhooks Tab
- **Webhook Management**: Create, edit, and delete webhooks
- **Supported Types**: Discord, Slack, Generic
- **Event Configuration**: Configure which events trigger webhooks
- **Enable/Disable**: Toggle webhooks on/off
- **Delivery History**: View webhook delivery status

#### Content Filters Tab
- **Filter Creation**: Create content filters to exclude/include episodes
- **Filter Types**: Title, Description, Duration, Regex
- **Toggle Filters**: Enable/disable filters without deletion
- **Pattern Matching**: Flexible pattern matching for content filtering

#### Preferences Tab
- **Audio Settings**:
  - Default audio format (M4A, MP3, Opus)
  - Audio quality selection
- **Feature Toggles**:
  - Enable/disable SponsorBlock
  - Auto-refresh podcasts
- **Customization**: Personalize your podcast experience

#### Backup & Restore Tab
- **Create Backups**:
  - Database-only backup
  - Full backup with audio files
- **Restore Functionality**: Restore from previous backups
- **Backup Management**:
  - View all available backups
  - Download backups
  - Delete old backups
  - Backup size and date information

### 5. Analytics Dashboard (`/admin/analytics`)
- **Summary Statistics**:
  - Total plays with trend indicators
  - Unique episodes played
  - Total listen time
  - Average episode length
- **Time-Series Charts**:
  - Plays over time
  - Listen time trends
- **Popular Content**:
  - Top 10 episodes ranking
  - Top podcasts by play count
- **Geographic Distribution**:
  - Country-based listener distribution
  - Visual representation with flags
  - Percentage breakdown
- **Engagement Metrics**:
  - Episode types breakdown (pie chart)
  - Peak listening hours
  - Overall engagement score (0-100)
- **Flexible Time Periods**: 24h, 7d, 30d, 90d, 1y

## Technical Architecture

### Frontend Stack

1. **HTML5**: Modern semantic markup
2. **Tailwind CSS**: Utility-first CSS framework (via CDN)
3. **Chart.js**: Beautiful, responsive charts
4. **Font Awesome**: Comprehensive icon library
5. **Vanilla JavaScript**: No framework dependencies

### File Structure

```
web/
├── static/
│   ├── css/
│   │   └── admin.css          # Custom styles and dark mode
│   └── js/
│       ├── api.js             # API client wrapper
│       └── utils.js           # Utility functions
├── admin/
│   ├── dashboard.html         # Main dashboard
│   ├── podcasts.html          # Podcast management
│   ├── episodes.html          # Episode browser
│   ├── settings.html          # Settings & configuration
│   └── analytics.html         # Analytics dashboard
└── templates/                 # Future template location
```

### API Integration

The web UI communicates with the backend through a comprehensive JavaScript API client (`/static/js/api.js`):

#### Supported Endpoints

**Podcasts**:
- `GET /search/podcasts` - Search and list podcasts
- `POST /api/batch/podcasts/add` - Batch add podcasts
- `POST /api/batch/refresh` - Batch refresh podcasts

**Episodes**:
- `GET /search/episodes` - Search episodes with filters
- `POST /api/batch/episodes/delete` - Batch delete episodes
- `GET /media/:videoId` - Stream episode audio

**Analytics**:
- `GET /api/analytics/dashboard` - Comprehensive dashboard data
- `GET /api/analytics/summary` - Summary statistics
- `GET /api/analytics/popular` - Popular episodes
- `GET /api/analytics/geographic` - Geographic distribution

**Webhooks**:
- `GET /api/webhooks` - List all webhooks
- `POST /api/webhooks` - Create webhook
- `PUT /api/webhooks/:id` - Update webhook
- `DELETE /api/webhooks/:id` - Delete webhook

**Filters**:
- `GET /api/filters` - List content filters
- `POST /api/filters` - Create filter
- `PATCH /api/filters/:id/toggle` - Toggle filter

**Backup**:
- `POST /api/backup/create` - Create backup
- `GET /api/backup/list` - List backups
- `POST /api/backup/restore` - Restore backup

**Batch Jobs**:
- `GET /api/batch/status/:jobId` - Check batch job status

### Design Features

#### Dark Mode
- Default dark theme optimized for long sessions
- Reduced eye strain with carefully chosen color palette
- Consistent dark mode across all pages

#### Responsive Design
- Mobile-first approach
- Tablet and desktop optimized layouts
- Collapsible sidebar on mobile
- Touch-friendly interface

#### User Experience
- **Loading States**: Spinners during data fetches
- **Toast Notifications**: Non-intrusive success/error messages
- **Confirmation Dialogs**: Prevent accidental destructive actions
- **Real-time Updates**: Auto-refresh capabilities
- **Error Handling**: Graceful error messages with details
- **Pagination**: Efficient browsing of large datasets
- **Search Debouncing**: Optimized search performance

#### Visual Elements
- **Stat Cards**: Gradient backgrounds with icons
- **Charts**: Interactive Chart.js visualizations
- **Tables**: Hover effects and row highlighting
- **Badges**: Color-coded status indicators
- **Icons**: Consistent Font Awesome icon usage
- **Smooth Transitions**: CSS transitions for better UX

## Utility Functions

### Data Formatting (`utils.js`)

- `formatDuration(seconds)` - Convert seconds to HH:MM:SS
- `formatBytes(bytes)` - Human-readable file sizes
- `formatDate(dateString)` - Relative and absolute dates
- `formatNumber(num)` - Number formatting with commas
- `truncateText(text, maxLength)` - Text truncation

### UI Helpers

- `showToast(message, type, duration)` - Toast notifications
- `confirmDialog(message, onConfirm, onCancel)` - Confirmation dialogs
- `showModal(title, content, buttons)` - Custom modals
- `showLoading(containerId)` - Loading indicators
- `debounce(func, wait)` - Debounce function calls

### Validation

- `validateYoutubeId(input)` - YouTube URL/ID validation
- `sanitizeHTML(html)` - XSS prevention

### Visual Helpers

- `getPodcastTypeBadge(type)` - Type badge HTML
- `getStatusBadge(status)` - Status badge HTML
- `createPagination(current, total, callback)` - Pagination HTML

## API Client Usage

### Basic Usage

```javascript
// Get all podcasts
const podcasts = await api.getPodcasts(20, 0);

// Search episodes
const episodes = await api.searchEpisodes({
  query: 'javascript',
  podcast_id: 'UC...',
  limit: 10
});

// Create backup
await api.createBackup(true); // with audio files

// Get analytics
const analytics = await api.getDashboardAnalytics('7d');
```

### Authentication

```javascript
// Set auth token (stored in localStorage)
api.setAuthToken('your-token-here');

// Get current token
const token = api.getAuthToken();

// Clear token
api.clearAuthToken();
```

### Error Handling

```javascript
try {
  await api.batchDeleteEpisodes(episodeIds);
  showToast('Episodes deleted successfully', 'success');
} catch (error) {
  showToast('Failed to delete: ' + parseError(error), 'error');
}
```

## Browser Support

- **Chrome/Edge**: 90+
- **Firefox**: 88+
- **Safari**: 14+
- **Mobile**: iOS Safari 14+, Chrome Android 90+

## Performance Considerations

1. **Lazy Loading**: Charts and data loaded on-demand
2. **Pagination**: Limits data transfer and rendering
3. **Debouncing**: Search inputs debounced to reduce API calls
4. **Caching**: API responses cached in browser (via API client)
5. **Minimal Dependencies**: Only essential libraries loaded

## Security Features

1. **XSS Prevention**: HTML sanitization for user content
2. **CSRF Protection**: Token-based authentication
3. **Input Validation**: Client-side validation before API calls
4. **Secure Defaults**: HTTPS enforcement ready
5. **Error Messages**: No sensitive data exposed in errors

## Customization

### Colors

Edit `web/static/css/admin.css` CSS variables:

```css
:root {
  --primary-color: #3b82f6;    /* Blue */
  --success-color: #10b981;    /* Green */
  --danger-color: #ef4444;     /* Red */
  --warning-color: #f59e0b;    /* Yellow */
  /* ... */
}
```

### Chart Themes

Modify Chart.js options in individual HTML files:

```javascript
{
  scales: {
    y: {
      grid: { color: '#334155' },  // Grid line color
      ticks: { color: '#94a3b8' }  // Label color
    }
  }
}
```

## Deployment Notes

### Routes Configured

The following routes have been added to `internal/app/controller.go`:

```go
// Static assets
e.Static("/static", "web/static")

// Admin pages
e.GET("/admin", ...)           // Dashboard
e.GET("/admin/podcasts", ...)  // Podcast management
e.GET("/admin/episodes", ...)  // Episode browser
e.GET("/admin/settings", ...)  // Settings
e.GET("/admin/analytics", ...) // Analytics
```

### Running the Application

1. Ensure the `web/` directory is in your project root
2. Start the Go application: `go run cmd/app/main.go`
3. Access the admin panel at: `http://localhost:8080/admin`

### Docker Deployment

If using Docker, ensure the web directory is included:

```dockerfile
COPY web/ /app/web/
```

## Future Enhancements

Potential improvements for future versions:

1. **Authentication UI**: Login/logout pages with JWT
2. **User Management**: Multi-user support with roles
3. **Real-time Updates**: WebSocket integration
4. **Advanced Analytics**: More detailed metrics and reports
5. **Theme Switcher**: Light/dark mode toggle
6. **Export Features**: CSV/PDF export of data
7. **Mobile App**: Progressive Web App (PWA) support
8. **Smart Playlists UI**: Visual playlist builder
9. **OPML Import/Export UI**: File upload interface
10. **Transcript Search**: Full-text search across transcripts

## Troubleshooting

### Common Issues

**1. Static files not loading**
- Check that the `web/` directory exists in the application root
- Verify file permissions
- Check browser console for 404 errors

**2. API calls failing**
- Check authentication token
- Verify backend is running
- Check CORS settings
- Review browser console for errors

**3. Charts not rendering**
- Ensure Chart.js CDN is accessible
- Check for JavaScript errors in console
- Verify data format matches chart expectations

**4. Styling issues**
- Verify Tailwind CSS CDN is loading
- Check custom CSS file is accessible
- Clear browser cache

## License

This web UI is part of the Clean Cast project and follows the same license.

## Credits

- **Tailwind CSS**: https://tailwindcss.com/
- **Chart.js**: https://www.chartjs.org/
- **Font Awesome**: https://fontawesome.com/
- **Echo Framework**: https://echo.labstack.com/

---

For more information, see the main project README or visit the project repository.

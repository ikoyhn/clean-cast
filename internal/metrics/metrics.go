package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCounter tracks HTTP requests by endpoint and method
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cleancast_http_requests_total",
			Help: "Total number of HTTP requests by endpoint and method",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestDuration tracks HTTP request duration
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cleancast_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// DownloadDuration tracks download duration for YouTube videos
	DownloadDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cleancast_download_duration_seconds",
			Help:    "Duration of YouTube video downloads in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1200}, // 1s to 20min
		},
	)

	// YouTubeAPIQuotaUsage tracks YouTube API quota usage
	YouTubeAPIQuotaUsage = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cleancast_youtube_api_quota_used_total",
			Help: "Total YouTube API quota units consumed",
		},
	)

	// YouTubeAPIRequests tracks YouTube API requests by operation
	YouTubeAPIRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cleancast_youtube_api_requests_total",
			Help: "Total number of YouTube API requests by operation",
		},
		[]string{"operation"},
	)

	// ActiveDownloads tracks currently active downloads
	ActiveDownloads = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cleancast_active_downloads",
			Help: "Number of currently active downloads",
		},
	)

	// ErrorCounter tracks errors by type
	ErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cleancast_errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type"},
	)

	// PodcastsServed tracks the number of podcasts served
	PodcastsServed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cleancast_podcasts_served_total",
			Help: "Total number of podcasts served by type (channel/playlist)",
		},
		[]string{"type"},
	)

	// MediaFilesServed tracks the number of media files served
	MediaFilesServed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cleancast_media_files_served_total",
			Help: "Total number of media files served",
		},
	)

	// SponsorBlockSegments tracks the number of SponsorBlock segments skipped
	SponsorBlockSegments = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cleancast_sponsorblock_segments_total",
			Help: "Total number of SponsorBlock segments processed",
		},
	)

	// DatabaseOperations tracks database operations
	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cleancast_database_operations_total",
			Help: "Total number of database operations by type",
		},
		[]string{"operation"},
	)
)

// RecordDownload records a download operation with duration
func RecordDownload(duration time.Duration) {
	DownloadDuration.Observe(duration.Seconds())
}

// RecordYouTubeAPICall records a YouTube API call and quota usage
func RecordYouTubeAPICall(operation string, quotaCost int) {
	YouTubeAPIRequests.WithLabelValues(operation).Inc()
	YouTubeAPIQuotaUsage.Add(float64(quotaCost))
}

// RecordError records an error by type
func RecordError(errorType string) {
	ErrorCounter.WithLabelValues(errorType).Inc()
}

// IncActiveDownloads increments the active downloads counter
func IncActiveDownloads() {
	ActiveDownloads.Inc()
}

// DecActiveDownloads decrements the active downloads counter
func DecActiveDownloads() {
	ActiveDownloads.Dec()
}

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	RequestCounter.WithLabelValues(method, endpoint, status).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

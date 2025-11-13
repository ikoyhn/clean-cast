package constants

// YouTube API pagination constants
const (
	// PageSize is the default page size for YouTube API requests
	PageSize = 50
)

// Database operation constants
const (
	// BatchSize is the number of records to insert in a single batch operation
	BatchSize = 100

	// CleanupDays is the number of days before old episode files are cleaned up
	CleanupDays = 7
)

// SponsorBlock constants
const (
	// SponsorBlockThreshold is the minimum difference in seconds that triggers a re-download
	// when sponsor segment timing changes
	SponsorBlockThreshold = 2.0
)

// Default configuration values
const (
	// DefaultMinDuration is the default minimum video duration to include in podcast feeds
	DefaultMinDuration = "3m"

	// DefaultPort is the default HTTP server port
	DefaultPort = "8082"

	// DefaultCron is the default cron schedule for cleanup jobs (weekly at midnight on Sunday)
	DefaultCron = "0 0 * * 0"
)

// Concurrency and rate limiting constants
const (
	// MaxConcurrentDownloads is the maximum number of concurrent video downloads
	MaxConcurrentDownloads = 5
)

// Cache and timeout constants
const (
	// RSSCacheTTL is the RSS feed cache time-to-live in minutes
	RSSCacheTTL = 15

	// RequestTimeout is the default HTTP request timeout in seconds
	RequestTimeout = 30
)

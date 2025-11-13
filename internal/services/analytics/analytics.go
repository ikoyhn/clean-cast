package analytics

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/http"
	"strings"
)

// Service provides analytics-related functionality
type Service struct{}

// NewService creates a new analytics service
func NewService() *Service {
	return &Service{}
}

// TrackEpisodePlay tracks when an episode is played
func (s *Service) TrackEpisodePlay(episodeId string, r *http.Request) error {
	ipAddress := getClientIP(r)
	country := getCountryFromIP(ipAddress)

	logger.Logger.Info().
		Str("episode_id", episodeId).
		Str("ip", ipAddress).
		Str("country", country).
		Msg("Tracking episode play")

	return database.TrackPlay(episodeId, ipAddress, country)
}

// TrackListenTime tracks the listen time for an episode based on HTTP range requests
func (s *Service) TrackListenTime(episodeId string, rangeHeader string, fileSize int64) error {
	listenTime := calculateListenTimeFromRange(rangeHeader, fileSize)

	if listenTime > 0 {
		logger.Logger.Debug().
			Str("episode_id", episodeId).
			Float64("listen_time", listenTime).
			Msg("Tracking listen time")

		return database.TrackListenTime(episodeId, listenTime)
	}

	return nil
}

// GetEpisodeAnalytics retrieves analytics for a specific episode
func (s *Service) GetEpisodeAnalytics(episodeId string) (*models.Analytics, error) {
	return database.GetEpisodeAnalytics(episodeId)
}

// GetPopularEpisodes retrieves the most popular episodes
func (s *Service) GetPopularEpisodes(limit int, days int) ([]models.Analytics, error) {
	if limit <= 0 {
		limit = 10
	}
	if days <= 0 {
		days = 7
	}

	return database.GetPopularEpisodes(limit, days)
}

// GetAnalyticsSummary retrieves overall analytics summary
func (s *Service) GetAnalyticsSummary() (map[string]interface{}, error) {
	return database.GetAnalyticsSummary()
}

// GetGeographicDistribution retrieves analytics grouped by country
func (s *Service) GetGeographicDistribution(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	return database.GetGeographicDistribution(limit)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// RemoteAddr includes port, strip it
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}

// getCountryFromIP determines the country from an IP address
// This is a placeholder - in production, you would use a GeoIP service/database
func getCountryFromIP(ip string) string {
	// TODO: Implement actual GeoIP lookup using MaxMind GeoLite2 or similar
	// For now, return empty string or implement basic logic

	// Check if it's a local/private IP
	if isLocalIP(ip) {
		return "LOCAL"
	}

	// In production, you would do something like:
	// geoIPDB, _ := geoip2.Open("/path/to/GeoLite2-Country.mmdb")
	// record, _ := geoIPDB.Country(net.ParseIP(ip))
	// return record.Country.IsoCode

	return "UNKNOWN"
}

// isLocalIP checks if an IP is a local/private IP
func isLocalIP(ip string) bool {
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	// Check for private IP ranges
	if strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.") {
		return true
	}

	return false
}

// calculateListenTimeFromRange estimates listen time from HTTP range requests
// This is an approximation based on the assumption that range requests indicate progressive playback
func calculateListenTimeFromRange(rangeHeader string, fileSize int64) float64 {
	if rangeHeader == "" || fileSize == 0 {
		return 0
	}

	// Parse range header (format: "bytes=start-end")
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeHeader, "-")

	if len(parts) != 2 {
		return 0
	}

	// For simplicity, we're not calculating exact listen time here
	// This would require tracking multiple range requests and correlating them
	// For now, we'll just return 0 and let the tracking happen at the /media endpoint level

	return 0
}

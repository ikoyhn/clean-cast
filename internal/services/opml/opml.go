package opml

import (
	"encoding/xml"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"regexp"
	"strings"
	"time"
)

// ParseOPML parses an OPML XML file and returns the structure
func ParseOPML(data []byte) (*models.OPML, error) {
	var opml models.OPML
	err := xml.Unmarshal(data, &opml)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to parse OPML XML")
		return nil, fmt.Errorf("failed to parse OPML: %w", err)
	}

	// Validate OPML version (support 1.0, 1.1, and 2.0)
	if opml.Version != "1.0" && opml.Version != "1.1" && opml.Version != "2.0" {
		logger.Logger.Warn().Str("version", opml.Version).Msg("Unsupported OPML version, attempting to parse anyway")
	}

	return &opml, nil
}

// ImportOPML imports podcast subscriptions from OPML
func ImportOPML(data []byte) *models.OPMLImportResponse {
	opml, err := ParseOPML(data)
	if err != nil {
		return &models.OPMLImportResponse{
			Success:      0,
			Failed:       1,
			TotalFeeds:   0,
			ErrorDetails: []string{err.Error()},
		}
	}

	response := &models.OPMLImportResponse{
		TotalFeeds:   len(opml.Body.Outlines),
		ErrorDetails: []string{},
	}

	for _, outline := range opml.Body.Outlines {
		// Skip non-RSS outlines or folders
		if outline.Type != "rss" || outline.XMLURL == "" {
			continue
		}

		// Try to extract YouTube channel/playlist ID from various URL formats
		channelId := extractYouTubeID(outline.XMLURL)
		if channelId == "" {
			// Try HTML URL as fallback
			channelId = extractYouTubeID(outline.HTMLURL)
		}

		if channelId == "" {
			response.Failed++
			response.ErrorDetails = append(response.ErrorDetails,
				fmt.Sprintf("Could not extract YouTube ID from feed: %s", outline.Text))
			continue
		}

		// Check if podcast already exists
		exists, err := database.PodcastExists(channelId)
		if err != nil {
			logger.Logger.Error().Err(err).Str("channelId", channelId).Msg("Error checking podcast existence")
			response.Failed++
			response.ErrorDetails = append(response.ErrorDetails,
				fmt.Sprintf("Database error for %s: %v", outline.Text, err))
			continue
		}

		if exists {
			logger.Logger.Info().Str("channelId", channelId).Msg("Podcast already exists, skipping")
			response.Success++
			continue
		}

		// Create a placeholder podcast entry
		// The actual data will be fetched when the RSS feed is requested
		podcast := &models.Podcast{
			Id:          channelId,
			PodcastName: outline.Title,
			Description: outline.Text,
		}

		database.SavePodcast(podcast)
		response.Success++
		logger.Logger.Info().Str("channelId", channelId).Str("name", outline.Title).Msg("Imported podcast from OPML")
	}

	return response
}

// ExportOPML exports CleanCast subscriptions to OPML format
func ExportOPML(host string) ([]byte, error) {
	// Get all podcasts from database
	podcasts, err := database.GetAllPodcasts()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to get podcasts for OPML export")
		return nil, fmt.Errorf("failed to get podcasts: %w", err)
	}

	// Create OPML structure
	opml := models.OPML{
		Version: "2.0",
		Head: models.OPMLHead{
			Title:        "CleanCast Subscriptions",
			DateCreated:  time.Now().Format(time.RFC1123),
			DateModified: time.Now().Format(time.RFC1123),
			OwnerName:    "CleanCast",
		},
		Body: models.OPMLBody{
			Outlines: []models.OPMLOutline{},
		},
	}

	// Add each podcast as an outline
	for _, podcast := range podcasts {
		xmlURL := fmt.Sprintf("%s/channel/%s", host, podcast.Id)
		htmlURL := fmt.Sprintf("https://www.youtube.com/channel/%s", podcast.Id)

		// Detect if it's a playlist based on ID format
		if strings.HasPrefix(podcast.Id, "PL") || strings.HasPrefix(podcast.Id, "UU") {
			xmlURL = fmt.Sprintf("%s/rss/%s", host, podcast.Id)
			htmlURL = fmt.Sprintf("https://www.youtube.com/playlist?list=%s", podcast.Id)
		}

		outline := models.OPMLOutline{
			Text:    podcast.PodcastName,
			Title:   podcast.PodcastName,
			Type:    "rss",
			XMLURL:  xmlURL,
			HTMLURL: htmlURL,
		}

		opml.Body.Outlines = append(opml.Body.Outlines, outline)
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(opml, "", "  ")
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to marshal OPML to XML")
		return nil, fmt.Errorf("failed to generate OPML: %w", err)
	}

	// Add XML header
	xmlHeader := []byte(xml.Header)
	result := append(xmlHeader, output...)

	return result, nil
}

// extractYouTubeID extracts YouTube channel or playlist ID from various URL formats
func extractYouTubeID(url string) string {
	if url == "" {
		return ""
	}

	// Pattern for channel IDs (UC... format)
	channelPattern := regexp.MustCompile(`(?:youtube\.com/channel/|/channel/)([a-zA-Z0-9_-]+)`)
	if matches := channelPattern.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	// Pattern for playlist IDs (PL... or UU... format)
	playlistPattern := regexp.MustCompile(`(?:youtube\.com/playlist\?list=|list=|/rss/)([a-zA-Z0-9_-]+)`)
	if matches := playlistPattern.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	// Pattern for custom URLs (@username format)
	customPattern := regexp.MustCompile(`youtube\.com/@([a-zA-Z0-9_-]+)`)
	if matches := customPattern.FindStringSubmatch(url); len(matches) > 1 {
		// Note: Custom URLs need to be resolved to channel IDs through YouTube API
		// For now, we'll just log a warning
		logger.Logger.Warn().Str("url", url).Msg("Custom YouTube URL detected, may need API resolution")
		return matches[1]
	}

	return ""
}

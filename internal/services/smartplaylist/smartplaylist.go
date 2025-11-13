package smartplaylist

import (
	"encoding/json"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/generator"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"strings"
	"time"
)

// CreateSmartPlaylist creates a new smart playlist
func CreateSmartPlaylist(req *models.SmartPlaylistCreateRequest) (*models.SmartPlaylist, error) {
	// Generate a unique ID
	id := generatePlaylistID(req.Name)

	// Encode rules to JSON
	rulesJSON, err := json.Marshal(models.SmartPlaylistRules{Rules: req.Rules})
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to encode smart playlist rules")
		return nil, fmt.Errorf("failed to encode rules: %w", err)
	}

	// Set default logic if not provided
	logic := req.Logic
	if logic == "" {
		logic = "AND"
	}

	playlist := &models.SmartPlaylist{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		Rules:       string(rulesJSON),
		Logic:       logic,
	}

	err = database.SaveSmartPlaylist(playlist)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

// UpdateSmartPlaylist updates an existing smart playlist
func UpdateSmartPlaylist(id string, req *models.SmartPlaylistUpdateRequest) (*models.SmartPlaylist, error) {
	playlist, err := database.GetSmartPlaylist(id)
	if err != nil {
		return nil, err
	}

	if playlist == nil {
		return nil, fmt.Errorf("smart playlist not found")
	}

	// Update fields if provided
	if req.Name != "" {
		playlist.Name = req.Name
	}
	if req.Description != "" {
		playlist.Description = req.Description
	}
	if req.Logic != "" {
		playlist.Logic = req.Logic
	}
	if req.Rules != nil && len(req.Rules) > 0 {
		rulesJSON, err := json.Marshal(models.SmartPlaylistRules{Rules: req.Rules})
		if err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to encode smart playlist rules")
			return nil, fmt.Errorf("failed to encode rules: %w", err)
		}
		playlist.Rules = string(rulesJSON)
	}

	err = database.UpdateSmartPlaylist(playlist)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

// DeleteSmartPlaylist deletes a smart playlist
func DeleteSmartPlaylist(id string) error {
	return database.DeleteSmartPlaylist(id)
}

// GetSmartPlaylist retrieves a smart playlist by ID
func GetSmartPlaylist(id string) (*models.SmartPlaylistResponse, error) {
	playlist, err := database.GetSmartPlaylist(id)
	if err != nil {
		return nil, err
	}

	if playlist == nil {
		return nil, fmt.Errorf("smart playlist not found")
	}

	return convertToResponse(playlist)
}

// GetAllSmartPlaylists retrieves all smart playlists
func GetAllSmartPlaylists() ([]models.SmartPlaylistResponse, error) {
	playlists, err := database.GetAllSmartPlaylists()
	if err != nil {
		return nil, err
	}

	responses := make([]models.SmartPlaylistResponse, len(playlists))
	for i, playlist := range playlists {
		response, err := convertToResponse(&playlist)
		if err != nil {
			logger.Logger.Error().Err(err).Str("id", playlist.Id).Msg("Failed to convert smart playlist to response")
			continue
		}
		responses[i] = *response
	}

	return responses, nil
}

// BuildSmartPlaylistRSSFeed generates an RSS feed for a smart playlist
func BuildSmartPlaylistRSSFeed(id string, host string) ([]byte, error) {
	playlist, err := database.GetSmartPlaylist(id)
	if err != nil {
		return nil, err
	}

	if playlist == nil {
		return nil, fmt.Errorf("smart playlist not found")
	}

	// Parse rules
	var rulesWrapper models.SmartPlaylistRules
	err = json.Unmarshal([]byte(playlist.Rules), &rulesWrapper)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to parse smart playlist rules")
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	// Get all episodes that match the rules
	episodes, err := getMatchingEpisodes(rulesWrapper.Rules, playlist.Logic)
	if err != nil {
		return nil, err
	}

	// Create a virtual podcast for the smart playlist
	virtualPodcast := models.Podcast{
		Id:            playlist.Id,
		PodcastName:   playlist.Name,
		Description:   playlist.Description,
		Category:      "Technology",
		ImageUrl:      "https://via.placeholder.com/1000x1000.png?text=Smart+Playlist",
		ArtistName:    "CleanCast Smart Playlist",
		Explicit:      "no",
		PodcastEpisodes: episodes,
	}

	// Generate RSS feed
	rssFeed := generateSmartPlaylistRSS(virtualPodcast, host)
	return rssFeed, nil
}

// getMatchingEpisodes retrieves episodes that match the smart playlist rules
func getMatchingEpisodes(rules []models.SmartPlaylistRule, logic string) ([]models.PodcastEpisode, error) {
	// Get all episodes from database
	allEpisodes, err := database.GetAllEpisodes()
	if err != nil {
		return nil, err
	}

	// Filter episodes based on rules
	var matchingEpisodes []models.PodcastEpisode
	for _, episode := range allEpisodes {
		if matchesRules(episode, rules, logic) {
			matchingEpisodes = append(matchingEpisodes, episode)
		}
	}

	return matchingEpisodes, nil
}

// matchesRules checks if an episode matches the smart playlist rules
func matchesRules(episode models.PodcastEpisode, rules []models.SmartPlaylistRule, logic string) bool {
	if len(rules) == 0 {
		return true
	}

	matchCount := 0
	for _, rule := range rules {
		if matchesRule(episode, rule) {
			matchCount++
			if logic == "OR" {
				return true // For OR logic, one match is enough
			}
		} else if logic == "AND" {
			return false // For AND logic, all must match
		}
	}

	// For AND logic, all rules must have matched
	return logic == "AND" && matchCount == len(rules)
}

// matchesRule checks if an episode matches a single rule
func matchesRule(episode models.PodcastEpisode, rule models.SmartPlaylistRule) bool {
	switch rule.Field {
	case "duration":
		return matchDuration(episode.Duration, rule.Operator, rule.Value)
	case "publish_date":
		return matchPublishDate(episode.PublishedDate, rule.Operator, rule.Value)
	case "keyword", "title":
		return matchKeyword(episode.EpisodeName, rule.Operator, rule.Value)
	case "description":
		return matchKeyword(episode.EpisodeDescription, rule.Operator, rule.Value)
	case "channel_id":
		return matchChannelID(episode.PodcastId, rule.Operator, rule.Value)
	default:
		logger.Logger.Warn().Str("field", rule.Field).Msg("Unknown rule field")
		return false
	}
}

// matchDuration checks if duration matches the rule
func matchDuration(duration time.Duration, operator string, value interface{}) bool {
	targetSeconds, ok := value.(float64)
	if !ok {
		return false
	}

	durationSeconds := duration.Seconds()

	switch operator {
	case "equals":
		return durationSeconds == targetSeconds
	case "greater_than":
		return durationSeconds > targetSeconds
	case "less_than":
		return durationSeconds < targetSeconds
	default:
		return false
	}
}

// matchPublishDate checks if publish date matches the rule
func matchPublishDate(publishDate time.Time, operator string, value interface{}) bool {
	targetDateStr, ok := value.(string)
	if !ok {
		return false
	}

	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		logger.Logger.Error().Err(err).Str("value", targetDateStr).Msg("Failed to parse target date")
		return false
	}

	switch operator {
	case "equals":
		return publishDate.Format("2006-01-02") == targetDate.Format("2006-01-02")
	case "before":
		return publishDate.Before(targetDate)
	case "after":
		return publishDate.After(targetDate)
	default:
		return false
	}
}

// matchKeyword checks if text matches the rule
func matchKeyword(text string, operator string, value interface{}) bool {
	keyword, ok := value.(string)
	if !ok {
		return false
	}

	lowerText := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)

	switch operator {
	case "equals":
		return lowerText == lowerKeyword
	case "contains":
		return strings.Contains(lowerText, lowerKeyword)
	default:
		return false
	}
}

// matchChannelID checks if channel ID matches the rule
func matchChannelID(channelID string, operator string, value interface{}) bool {
	targetID, ok := value.(string)
	if !ok {
		return false
	}

	switch operator {
	case "equals":
		return channelID == targetID
	case "contains":
		return strings.Contains(channelID, targetID)
	default:
		return false
	}
}

// generateSmartPlaylistRSS generates RSS feed for smart playlist
func generateSmartPlaylistRSS(podcast models.Podcast, host string) []byte {
	logger.Logger.Info().Msg("[SMART PLAYLIST] Generating RSS Feed...")

	podcastLink := fmt.Sprintf("%s/rss/smart/%s", host, podcast.Id)

	now := time.Now()
	ytPodcast := generator.New(podcast.PodcastName, podcastLink, podcast.Description, &now)
	ytPodcast.AddImage(podcast.ImageUrl)
	ytPodcast.AddCategory(podcast.Category, []string{""})
	ytPodcast.Docs = "http://www.rssboard.org/rss-specification"
	ytPodcast.IAuthor = podcast.ArtistName

	for _, episode := range podcast.PodcastEpisodes {
		if episode.EpisodeName == "Private video" || episode.EpisodeDescription == "This video is private." {
			continue
		}

		mediaUrl := fmt.Sprintf("%s/media/%s.m4a", host, episode.YoutubeVideoId)

		enclosure := generator.Enclosure{
			URL:    mediaUrl,
			Length: 0,
			Type:   generator.M4A,
		}

		podcastItem := generator.Item{
			Title:       episode.EpisodeName,
			Description: episode.EpisodeDescription,
			GUID: struct {
				Value       string `xml:",chardata"`
				IsPermaLink bool   `xml:"isPermaLink,attr"`
			}{
				Value:       episode.YoutubeVideoId,
				IsPermaLink: false,
			},
			Enclosure: &enclosure,
			PubDate:   &episode.PublishedDate,
		}
		ytPodcast.AddItem(podcastItem)
	}

	return ytPodcast.Bytes()
}

// convertToResponse converts a SmartPlaylist model to a response
func convertToResponse(playlist *models.SmartPlaylist) (*models.SmartPlaylistResponse, error) {
	var rulesWrapper models.SmartPlaylistRules
	err := json.Unmarshal([]byte(playlist.Rules), &rulesWrapper)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to parse smart playlist rules")
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	return &models.SmartPlaylistResponse{
		Id:          playlist.Id,
		Name:        playlist.Name,
		Description: playlist.Description,
		Rules:       rulesWrapper.Rules,
		Logic:       playlist.Logic,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}, nil
}

// generatePlaylistID generates a unique ID for a playlist based on name
func generatePlaylistID(name string) string {
	// Create a simple ID based on timestamp and name
	timestamp := time.Now().Unix()
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	cleanName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, cleanName)

	return fmt.Sprintf("smart-%d-%s", timestamp, cleanName)
}

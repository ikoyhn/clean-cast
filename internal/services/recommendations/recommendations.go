package recommendations

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RecommendationService provides podcast and episode recommendations
type RecommendationService struct {
	db *gorm.DB
}

// NewRecommendationService creates a new recommendation service instance
func NewRecommendationService(db *gorm.DB) *RecommendationService {
	return &RecommendationService{db: db}
}

// SimilarPodcast represents a similar podcast with a similarity score
type SimilarPodcast struct {
	Podcast         *models.Podcast `json:"podcast"`
	SimilarityScore float64         `json:"similarity_score"`
}

// TrendingEpisode represents a trending episode with play count
type TrendingEpisode struct {
	Episode   *models.PodcastEpisode `json:"episode"`
	PlayCount int64                  `json:"play_count"`
}

// GetSimilarPodcasts finds podcasts similar to the given podcast ID
// based on category and keywords in the description
func (s *RecommendationService) GetSimilarPodcasts(podcastId string, limit int) ([]SimilarPodcast, error) {
	// Get the source podcast
	sourcePodcast := database.GetPodcast(podcastId)
	if sourcePodcast == nil {
		return nil, nil
	}

	// Get all podcasts
	var allPodcasts []models.Podcast
	err := s.db.Find(&allPodcasts).Error
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to fetch podcasts for similarity comparison")
		return nil, err
	}

	// Calculate similarity scores
	var similarPodcasts []SimilarPodcast
	for _, podcast := range allPodcasts {
		// Skip the source podcast itself
		if podcast.Id == podcastId {
			continue
		}

		score := calculateSimilarity(sourcePodcast, &podcast)
		if score > 0 {
			similarPodcasts = append(similarPodcasts, SimilarPodcast{
				Podcast:         &podcast,
				SimilarityScore: score,
			})
		}
	}

	// Sort by similarity score (bubble sort for simplicity)
	for i := 0; i < len(similarPodcasts)-1; i++ {
		for j := 0; j < len(similarPodcasts)-i-1; j++ {
			if similarPodcasts[j].SimilarityScore < similarPodcasts[j+1].SimilarityScore {
				similarPodcasts[j], similarPodcasts[j+1] = similarPodcasts[j+1], similarPodcasts[j]
			}
		}
	}

	// Return top N results
	if len(similarPodcasts) > limit {
		similarPodcasts = similarPodcasts[:limit]
	}

	return similarPodcasts, nil
}

// GetTrendingEpisodes returns the most played episodes in the last 7 days
func (s *RecommendationService) GetTrendingEpisodes(limit int) ([]TrendingEpisode, error) {
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

	// Get playback history from the last 7 days
	var histories []models.EpisodePlaybackHistory
	err := s.db.Where("last_access_date >= ?", sevenDaysAgo).
		Find(&histories).Error
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to fetch trending episodes")
		return nil, err
	}

	// Count plays per episode
	playCounts := make(map[string]int64)
	for _, history := range histories {
		playCounts[history.YoutubeVideoId]++
	}

	// Get episode details
	var trendingEpisodes []TrendingEpisode
	for videoId, count := range playCounts {
		var episode models.PodcastEpisode
		err := s.db.Where("youtube_video_id = ?", videoId).First(&episode).Error
		if err != nil {
			logger.Logger.Warn().
				Err(err).
				Str("video_id", videoId).
				Msg("Episode not found in database")
			continue
		}

		trendingEpisodes = append(trendingEpisodes, TrendingEpisode{
			Episode:   &episode,
			PlayCount: count,
		})
	}

	// Sort by play count (bubble sort)
	for i := 0; i < len(trendingEpisodes)-1; i++ {
		for j := 0; j < len(trendingEpisodes)-i-1; j++ {
			if trendingEpisodes[j].PlayCount < trendingEpisodes[j+1].PlayCount {
				trendingEpisodes[j], trendingEpisodes[j+1] = trendingEpisodes[j+1], trendingEpisodes[j]
			}
		}
	}

	// Return top N results
	if len(trendingEpisodes) > limit {
		trendingEpisodes = trendingEpisodes[:limit]
	}

	return trendingEpisodes, nil
}

// GetRelatedEpisodes finds episodes from the same channel/playlist as the given video
func (s *RecommendationService) GetRelatedEpisodes(videoId string, limit int) ([]models.PodcastEpisode, error) {
	// Get the source episode
	var sourceEpisode models.PodcastEpisode
	err := s.db.Where("youtube_video_id = ?", videoId).First(&sourceEpisode).Error
	if err != nil {
		logger.Logger.Error().
			Err(err).
			Str("video_id", videoId).
			Msg("Source episode not found")
		return nil, err
	}

	// Get related episodes from the same podcast
	var relatedEpisodes []models.PodcastEpisode
	err = s.db.Where("podcast_id = ? AND youtube_video_id != ?", sourceEpisode.PodcastId, videoId).
		Order("published_date DESC").
		Limit(limit).
		Find(&relatedEpisodes).Error
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to fetch related episodes")
		return nil, err
	}

	return relatedEpisodes, nil
}

// GetPersonalizedRecommendations returns personalized recommendations based on user's listening history
func (s *RecommendationService) GetPersonalizedRecommendations(limit int) ([]models.PodcastEpisode, error) {
	// Get recent playback history (last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour).Unix()

	var histories []models.EpisodePlaybackHistory
	err := s.db.Where("last_access_date >= ?", thirtyDaysAgo).
		Order("last_access_date DESC").
		Limit(100). // Get top 100 recently played
		Find(&histories).Error
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to fetch playback history")
		return nil, err
	}

	if len(histories) == 0 {
		// No history, return trending episodes instead
		trending, err := s.GetTrendingEpisodes(limit)
		if err != nil {
			return nil, err
		}

		var episodes []models.PodcastEpisode
		for _, t := range trending {
			episodes = append(episodes, *t.Episode)
		}
		return episodes, nil
	}

	// Track which podcasts the user has listened to
	podcastFrequency := make(map[string]int)
	listenedVideoIds := make(map[string]bool)

	for _, history := range histories {
		var episode models.PodcastEpisode
		err := s.db.Where("youtube_video_id = ?", history.YoutubeVideoId).First(&episode).Error
		if err == nil {
			podcastFrequency[episode.PodcastId]++
			listenedVideoIds[episode.YoutubeVideoId] = true
		}
	}

	// Find most listened podcast
	var topPodcastId string
	maxFreq := 0
	for podcastId, freq := range podcastFrequency {
		if freq > maxFreq {
			maxFreq = freq
			topPodcastId = podcastId
		}
	}

	// Get recommendations from the top podcast and similar podcasts
	var recommendations []models.PodcastEpisode

	// Get unlistened episodes from favorite podcast
	if topPodcastId != "" {
		var podcastEpisodes []models.PodcastEpisode
		err = s.db.Where("podcast_id = ?", topPodcastId).
			Order("published_date DESC").
			Limit(limit).
			Find(&podcastEpisodes).Error
		if err == nil {
			for _, ep := range podcastEpisodes {
				if !listenedVideoIds[ep.YoutubeVideoId] {
					recommendations = append(recommendations, ep)
					if len(recommendations) >= limit {
						return recommendations, nil
					}
				}
			}
		}

		// Get episodes from similar podcasts
		similarPodcasts, err := s.GetSimilarPodcasts(topPodcastId, 5)
		if err == nil {
			for _, similar := range similarPodcasts {
				var episodes []models.PodcastEpisode
				err = s.db.Where("podcast_id = ?", similar.Podcast.Id).
					Order("published_date DESC").
					Limit(5).
					Find(&episodes).Error
				if err == nil {
					for _, ep := range episodes {
						if !listenedVideoIds[ep.YoutubeVideoId] {
							recommendations = append(recommendations, ep)
							if len(recommendations) >= limit {
								return recommendations[:limit], nil
							}
						}
					}
				}
			}
		}
	}

	return recommendations, nil
}

// calculateSimilarity calculates a similarity score between two podcasts
// based on category match and keyword overlap in descriptions
func calculateSimilarity(podcast1, podcast2 *models.Podcast) float64 {
	score := 0.0

	// Category match (weight: 50%)
	if podcast1.Category != "" && podcast1.Category == podcast2.Category {
		score += 0.5
	}

	// Description keyword overlap (weight: 50%)
	keywords1 := extractKeywords(podcast1.Description)
	keywords2 := extractKeywords(podcast2.Description)

	if len(keywords1) > 0 && len(keywords2) > 0 {
		matchCount := 0
		for keyword := range keywords1 {
			if keywords2[keyword] {
				matchCount++
			}
		}

		// Calculate Jaccard similarity
		union := len(keywords1) + len(keywords2) - matchCount
		if union > 0 {
			score += 0.5 * (float64(matchCount) / float64(union))
		}
	}

	return score
}

// extractKeywords extracts meaningful keywords from a text
func extractKeywords(text string) map[string]bool {
	// Common stop words to exclude
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"up": true, "about": true, "into": true, "through": true, "during": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "i": true,
		"you": true, "he": true, "she": true, "it": true, "we": true,
		"they": true, "what": true, "which": true, "who": true, "when": true,
		"where": true, "why": true, "how": true, "all": true, "each": true,
		"every": true, "both": true, "few": true, "more": true, "most": true,
		"other": true, "some": true, "such": true, "no": true, "nor": true,
		"not": true, "only": true, "own": true, "same": true, "so": true,
		"than": true, "too": true, "very": true, "can": true, "just": true,
		"as": true, "if": true, "out": true, "over": true, "then": true,
	}

	keywords := make(map[string]bool)
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}—")

		// Only include words longer than 3 characters that aren't stop words
		if len(word) > 3 && !stopWords[word] {
			keywords[word] = true
		}
	}

	return keywords
}

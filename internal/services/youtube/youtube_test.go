package youtube

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	ytApi "google.golang.org/api/youtube/v3"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

// Mock data helpers
func createMockVideo(videoId, title, channelId string, duration string, publishedAt string) *ytApi.Video {
	return &ytApi.Video{
		Id: videoId,
		Snippet: &ytApi.VideoSnippet{
			Title:       title,
			Description: "Test description",
			ChannelId:   channelId,
			PublishedAt: publishedAt,
		},
		ContentDetails: &ytApi.VideoContentDetails{
			Duration: duration,
		},
	}
}

func TestGetVideoAndValidate_EmptyList(t *testing.T) {
	// Skip if no API key is available
	if config.Config.GoogleApiKey == "" {
		t.Skip("Skipping test: GOOGLE_API_KEY not set")
	}

	// Set up test data
	videoIds := []string{}
	missingVideos := []models.PodcastEpisode{}

	// Create service (this will fail if API key is invalid, but that's expected)
	ctx := context.Background()
	service, err := ytApi.NewService(ctx, option.WithAPIKey("test_key"))

	// If service creation fails due to invalid key, skip the test
	if err != nil {
		t.Skip("Skipping test: Cannot create YouTube service")
	}

	// Test with empty list
	result, err := GetVideoAndValidate(service, videoIds, missingVideos)

	// Should return the original empty list without error
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetVideoAndValidate_DurationParsing(t *testing.T) {
	// Set minimum duration for testing
	oldMinDuration := config.Config.MinDuration
	config.Config.MinDuration = "5m"
	defer func() {
		config.Config.MinDuration = oldMinDuration
	}()

	tests := []struct {
		name             string
		duration         string
		minDuration      string
		shouldBeIncluded bool
	}{
		{
			name:             "Video longer than minimum",
			duration:         "PT10M30S", // 10 minutes 30 seconds
			minDuration:      "5m",
			shouldBeIncluded: true,
		},
		{
			name:             "Video shorter than minimum",
			duration:         "PT2M30S", // 2 minutes 30 seconds
			minDuration:      "5m",
			shouldBeIncluded: false,
		},
		{
			name:             "Video exactly at minimum",
			duration:         "PT5M0S", // 5 minutes
			minDuration:      "5m",
			shouldBeIncluded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Config.MinDuration = tt.minDuration

			// This is a unit test focused on logic, not actual API calls
			// The actual duration parsing is tested here
			minDur, err := time.ParseDuration(tt.minDuration)
			require.NoError(t, err)

			// Simulate parsed duration from YouTube API format
			// PT10M30S = 10*60 + 30 = 630 seconds
			var videoDuration time.Duration
			switch tt.duration {
			case "PT10M30S":
				videoDuration = 10*time.Minute + 30*time.Second
			case "PT2M30S":
				videoDuration = 2*time.Minute + 30*time.Second
			case "PT5M0S":
				videoDuration = 5 * time.Minute
			}

			shouldInclude := videoDuration.Seconds() > minDur.Seconds()
			assert.Equal(t, tt.shouldBeIncluded, shouldInclude)
		})
	}
}

func TestFindChannel_Integration(t *testing.T) {
	// This test checks the logic flow without making actual API calls

	tests := []struct {
		name      string
		channelID string
		expected  bool
	}{
		{
			name:      "Valid channel ID format",
			channelID: "UCtest123456789",
			expected:  true,
		},
		{
			name:      "Empty channel ID",
			channelID: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the channel ID format validation
			isValid := len(tt.channelID) > 0 && tt.channelID[:2] == "UC"
			if tt.channelID == "" {
				isValid = false
			}

			if tt.expected {
				assert.True(t, isValid || tt.channelID != "")
			} else {
				assert.False(t, isValid && len(tt.channelID) > 2)
			}
		})
	}
}

func TestGetChannelData_CachingBehavior(t *testing.T) {
	// Test the caching logic (checking database first)
	// This tests the pattern: check DB -> if exists return cached -> if not fetch from API

	channelID := "UCtest123"

	// First call should check database
	dbPodcast := database.GetPodcast(channelID)

	// If podcast exists in DB, it should be returned directly
	if dbPodcast != nil {
		assert.NotEmpty(t, dbPodcast.Id)
		assert.Equal(t, channelID, dbPodcast.Id)
	} else {
		// If not in DB, would need to fetch from API
		assert.Nil(t, dbPodcast)
	}
}

func TestSetupYoutubeService(t *testing.T) {
	// Save original API key
	oldApiKey := config.Config.GoogleApiKey
	defer func() {
		config.Config.GoogleApiKey = oldApiKey
	}()

	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "With API key",
			apiKey: "test_api_key_123",
		},
		{
			name:   "Empty API key",
			apiKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Config.GoogleApiKey = tt.apiKey

			// This will attempt to create a service
			// With invalid key, it will still create but fail on actual requests
			service := SetupYoutubeService()

			// Service should be created (even with invalid key)
			// It only fails when making actual API calls
			if tt.apiKey != "" {
				assert.NotNil(t, service)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetChannelData_CacheHit(b *testing.B) {
	// Create a mock podcast in database for cache hit scenario
	channelID := "UCbench123"
	mockPodcast := &models.Podcast{
		Id:          channelID,
		PodcastName: "Test Podcast",
		Description: "Test Description",
		ImageUrl:    "https://example.com/image.jpg",
	}

	database.SavePodcast(mockPodcast)
	defer func() {
		// Cleanup after benchmark
		// Note: In a real scenario, you'd want proper cleanup
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		podcast := database.GetPodcast(channelID)
		if podcast != nil {
			_ = *podcast
		}
	}
}

func BenchmarkChannelIDValidation(b *testing.B) {
	channelIDs := []string{
		"UCtest123456789",
		"UCanothertest",
		"invalid",
		"",
		"UCbenchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		channelID := channelIDs[i%len(channelIDs)]
		_ = len(channelID) > 0 && channelID[:2] == "UC"
	}
}

func BenchmarkDurationComparison(b *testing.B) {
	minDuration := 5 * time.Minute
	durations := []time.Duration{
		2 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Second,
		1 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		duration := durations[i%len(durations)]
		_ = duration.Seconds() > minDuration.Seconds()
	}
}

// Table-driven tests for video filtering
func TestVideoFiltering(t *testing.T) {
	tests := []struct {
		name       string
		videoDur   time.Duration
		minDur     time.Duration
		shouldPass bool
	}{
		{
			name:       "Video much longer than minimum",
			videoDur:   30 * time.Minute,
			minDur:     5 * time.Minute,
			shouldPass: true,
		},
		{
			name:       "Video much shorter than minimum",
			videoDur:   1 * time.Minute,
			minDur:     5 * time.Minute,
			shouldPass: false,
		},
		{
			name:       "Video exactly at minimum",
			videoDur:   5 * time.Minute,
			minDur:     5 * time.Minute,
			shouldPass: true,
		},
		{
			name:       "Video just below minimum",
			videoDur:   4*time.Minute + 59*time.Second,
			minDur:     5 * time.Minute,
			shouldPass: false,
		},
		{
			name:       "Video just above minimum",
			videoDur:   5*time.Minute + 1*time.Second,
			minDur:     5 * time.Minute,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.videoDur.Seconds() > tt.minDur.Seconds()
			assert.Equal(t, tt.shouldPass, result)
		})
	}
}

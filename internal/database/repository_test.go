package database

import (
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ytApi "google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = testDB.AutoMigrate(
		&models.Podcast{},
		&models.PodcastEpisode{},
		&models.EpisodePlaybackHistory{},
	)
	require.NoError(t, err)

	// Set the global db variable for the package
	db = testDB

	return testDB
}

func TestPodcastExists(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		podcastId   string
		shouldExist bool
		setupFunc   func()
	}{
		{
			name:        "Podcast does not exist",
			podcastId:   "UC_nonexistent",
			shouldExist: false,
			setupFunc:   func() {},
		},
		{
			name:        "Podcast exists",
			podcastId:   "UC_existing",
			shouldExist: true,
			setupFunc: func() {
				podcast := &models.Podcast{
					Id:          "UC_existing",
					PodcastName: "Test Podcast",
					Description: "Test Description",
				}
				db.Create(podcast)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDB(t) // Fresh DB for each test
			tt.setupFunc()

			exists, err := PodcastExists(tt.podcastId)
			assert.NoError(t, err)
			assert.Equal(t, tt.shouldExist, exists)
		})
	}
}

func TestGetPodcast(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name      string
		podcastId string
		setupFunc func()
		expected  *models.Podcast
	}{
		{
			name:      "Podcast not found",
			podcastId: "UC_notfound",
			setupFunc: func() {},
			expected:  nil,
		},
		{
			name:      "Podcast found",
			podcastId: "UC_found",
			setupFunc: func() {
				podcast := &models.Podcast{
					Id:          "UC_found",
					PodcastName: "Found Podcast",
					Description: "Found Description",
					ImageUrl:    "https://example.com/image.jpg",
				}
				db.Create(podcast)
			},
			expected: &models.Podcast{
				Id:          "UC_found",
				PodcastName: "Found Podcast",
				Description: "Found Description",
				ImageUrl:    "https://example.com/image.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDB(t)
			tt.setupFunc()

			result := GetPodcast(tt.podcastId)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Id, result.Id)
				assert.Equal(t, tt.expected.PodcastName, result.PodcastName)
				assert.Equal(t, tt.expected.Description, result.Description)
			}
		})
	}
}

func TestSavePodcast(t *testing.T) {
	setupTestDB(t)

	podcast := &models.Podcast{
		Id:          "UC_newsave",
		PodcastName: "New Podcast",
		Description: "New Description",
		ImageUrl:    "https://example.com/new.jpg",
		ArtistName:  "Test Artist",
		Explicit:    "false",
	}

	SavePodcast(podcast)

	// Verify it was saved
	result := GetPodcast(podcast.Id)
	require.NotNil(t, result)
	assert.Equal(t, podcast.Id, result.Id)
	assert.Equal(t, podcast.PodcastName, result.PodcastName)
}

func TestEpisodeExists(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name           string
		youtubeVideoId string
		episodeType    string
		shouldExist    bool
		setupFunc      func()
	}{
		{
			name:           "Episode does not exist",
			youtubeVideoId: "video_notexist",
			episodeType:    "CHANNEL",
			shouldExist:    false,
			setupFunc:      func() {},
		},
		{
			name:           "Episode exists",
			youtubeVideoId: "video_exists",
			episodeType:    "CHANNEL",
			shouldExist:    true,
			setupFunc: func() {
				episode := &models.PodcastEpisode{
					YoutubeVideoId:     "video_exists",
					EpisodeName:        "Test Episode",
					EpisodeDescription: "Test Description",
					Type:               "CHANNEL",
					PodcastId:          "UC_test",
					PublishedDate:      time.Now(),
				}
				db.Create(episode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestDB(t)
			tt.setupFunc()

			exists, err := EpisodeExists(tt.youtubeVideoId, tt.episodeType)
			assert.NoError(t, err)
			assert.Equal(t, tt.shouldExist, exists)
		})
	}
}

func TestGetLatestEpisode(t *testing.T) {
	setupTestDB(t)

	podcastId := "UC_latest_test"

	// Create multiple episodes with different dates
	episodes := []models.PodcastEpisode{
		{
			YoutubeVideoId: "video1",
			EpisodeName:    "Episode 1",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-48 * time.Hour),
		},
		{
			YoutubeVideoId: "video2",
			EpisodeName:    "Episode 2",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-24 * time.Hour),
		},
		{
			YoutubeVideoId: "video3",
			EpisodeName:    "Latest Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now(),
		},
	}

	for _, ep := range episodes {
		db.Create(&ep)
	}

	// Get latest episode
	latest, err := GetLatestEpisode(podcastId)
	assert.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, "video3", latest.YoutubeVideoId)
	assert.Equal(t, "Latest Episode", latest.EpisodeName)
}

func TestGetOldestEpisode(t *testing.T) {
	setupTestDB(t)

	podcastId := "UC_oldest_test"

	// Create multiple episodes
	episodes := []models.PodcastEpisode{
		{
			YoutubeVideoId: "video1",
			EpisodeName:    "Oldest Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-72 * time.Hour),
		},
		{
			YoutubeVideoId: "video2",
			EpisodeName:    "Middle Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-24 * time.Hour),
		},
		{
			YoutubeVideoId: "video3",
			EpisodeName:    "Newest Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now(),
		},
	}

	for _, ep := range episodes {
		db.Create(&ep)
	}

	// Get oldest episode
	oldest, err := GetOldestEpisode(podcastId)
	assert.NoError(t, err)
	require.NotNil(t, oldest)
	assert.Equal(t, "video1", oldest.YoutubeVideoId)
	assert.Equal(t, "Oldest Episode", oldest.EpisodeName)
}

func TestGetAllPodcastEpisodeIds(t *testing.T) {
	setupTestDB(t)

	podcastId := "UC_all_test"

	// Create episodes
	expectedIds := []string{"video1", "video2", "video3"}
	for _, id := range expectedIds {
		episode := &models.PodcastEpisode{
			YoutubeVideoId: id,
			EpisodeName:    "Episode " + id,
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now(),
		}
		db.Create(episode)
	}

	// Get all IDs
	ids, err := GetAllPodcastEpisodeIds(podcastId)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedIds, ids)
}

func TestIsEpisodeSaved(t *testing.T) {
	setupTestDB(t)

	// Create a saved episode
	episode := &models.PodcastEpisode{
		YoutubeVideoId: "saved_video",
		EpisodeName:    "Saved Episode",
		Type:           "CHANNEL",
		PodcastId:      "UC_test",
		PublishedDate:  time.Now(),
	}
	db.Create(episode)

	tests := []struct {
		name     string
		videoId  string
		expected bool
	}{
		{
			name:     "Episode is saved",
			videoId:  "saved_video",
			expected: true,
		},
		{
			name:     "Episode is not saved",
			videoId:  "not_saved",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVideo := &ytApi.Video{
				Id: tt.videoId,
			}
			result := IsEpisodeSaved(mockVideo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPodcastEpisodesByPodcastId(t *testing.T) {
	setupTestDB(t)

	// Set up config
	config.Config.MinDuration = "5m"

	podcastId := "UC_get_episodes"

	// Create episodes with different durations
	episodes := []models.PodcastEpisode{
		{
			YoutubeVideoId: "video1",
			EpisodeName:    "Long Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			Duration:       10 * time.Minute,
			PublishedDate:  time.Now(),
		},
		{
			YoutubeVideoId: "video2",
			EpisodeName:    "Short Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			Duration:       2 * time.Minute,
			PublishedDate:  time.Now().Add(-24 * time.Hour),
		},
	}

	for _, ep := range episodes {
		db.Create(&ep)
	}

	// Test for CHANNEL type (filters by duration)
	result, err := GetPodcastEpisodesByPodcastId(podcastId, enum.CHANNEL)
	assert.NoError(t, err)
	assert.Len(t, result, 1) // Only long episode should be returned
	assert.Equal(t, "video1", result[0].YoutubeVideoId)
}

func TestSavePlaylistEpisodes(t *testing.T) {
	setupTestDB(t)

	episodes := []models.PodcastEpisode{
		{
			YoutubeVideoId: "pl_video1",
			EpisodeName:    "Playlist Episode 1",
			Type:           "PLAYLIST",
			PodcastId:      "PL_test",
			PublishedDate:  time.Now(),
		},
		{
			YoutubeVideoId: "pl_video2",
			EpisodeName:    "Playlist Episode 2",
			Type:           "PLAYLIST",
			PodcastId:      "PL_test",
			PublishedDate:  time.Now(),
		},
	}

	SavePlaylistEpisodes(episodes)

	// Verify episodes were saved
	var count int64
	db.Model(&models.PodcastEpisode{}).Where("podcast_id = ?", "PL_test").Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestDeletePodcastCronJob(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cleanup_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	setupTestDB(t)

	// Set config
	oldAudioDir := config.Config.AudioDir
	config.Config.AudioDir = tempDir + "/"
	defer func() {
		config.Config.AudioDir = oldAudioDir
	}()

	// Create old playback history
	oldHistory := &models.EpisodePlaybackHistory{
		YoutubeVideoId:   "old_video",
		LastAccessDate:   time.Now().Add(-8 * 24 * time.Hour).Unix(),
		TotalTimeSkipped: 10.5,
	}
	db.Create(oldHistory)

	// Create recent playback history
	recentHistory := &models.EpisodePlaybackHistory{
		YoutubeVideoId:   "recent_video",
		LastAccessDate:   time.Now().Unix(),
		TotalTimeSkipped: 5.5,
	}
	db.Create(recentHistory)

	// Create corresponding files
	os.WriteFile(tempDir+"/old_video.m4a", []byte("old content"), 0644)
	os.WriteFile(tempDir+"/recent_video.m4a", []byte("recent content"), 0644)

	// Run cleanup
	DeletePodcastCronJob()

	// Verify old history was deleted
	var oldHistoryCheck models.EpisodePlaybackHistory
	err = db.Where("youtube_video_id = ?", "old_video").First(&oldHistoryCheck).Error
	assert.Error(t, err) // Should not be found

	// Verify recent history still exists
	var recentHistoryCheck models.EpisodePlaybackHistory
	err = db.Where("youtube_video_id = ?", "recent_video").First(&recentHistoryCheck).Error
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkPodcastExists(b *testing.B) {
	testDB := setupBenchDB(b)
	db = testDB

	// Create test podcast
	podcast := &models.Podcast{
		Id:          "UC_bench",
		PodcastName: "Benchmark Podcast",
	}
	db.Create(podcast)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PodcastExists("UC_bench")
	}
}

func BenchmarkGetPodcast(b *testing.B) {
	testDB := setupBenchDB(b)
	db = testDB

	podcast := &models.Podcast{
		Id:          "UC_get_bench",
		PodcastName: "Get Benchmark",
		Description: "Description",
	}
	db.Create(podcast)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPodcast("UC_get_bench")
	}
}

func BenchmarkEpisodeExists(b *testing.B) {
	testDB := setupBenchDB(b)
	db = testDB

	episode := &models.PodcastEpisode{
		YoutubeVideoId: "bench_video",
		EpisodeName:    "Bench Episode",
		Type:           "CHANNEL",
		PodcastId:      "UC_test",
		PublishedDate:  time.Now(),
	}
	db.Create(episode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EpisodeExists("bench_video", "CHANNEL")
	}
}

func BenchmarkGetLatestEpisode(b *testing.B) {
	testDB := setupBenchDB(b)
	db = testDB

	podcastId := "UC_latest_bench"
	for i := 0; i < 10; i++ {
		episode := &models.PodcastEpisode{
			YoutubeVideoId: "video" + string(rune(i)),
			EpisodeName:    "Episode",
			PodcastId:      podcastId,
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
		db.Create(episode)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetLatestEpisode(podcastId)
	}
}

// Helper function for benchmarks
func setupBenchDB(b *testing.B) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(b, err)

	err = testDB.AutoMigrate(
		&models.Podcast{},
		&models.PodcastEpisode{},
		&models.EpisodePlaybackHistory{},
	)
	require.NoError(b, err)

	return testDB
}

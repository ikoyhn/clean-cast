package rss

import (
	"testing"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
)

// Benchmark RSS feed generation with various sizes
func BenchmarkGenerateRssFeed_SmallFeed(b *testing.B) {
	podcast := createTestPodcast(10)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

func BenchmarkGenerateRssFeed_MediumFeed(b *testing.B) {
	podcast := createTestPodcast(50)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

func BenchmarkGenerateRssFeed_LargeFeed(b *testing.B) {
	podcast := createTestPodcast(100)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

func BenchmarkGenerateRssFeed_VeryLargeFeed(b *testing.B) {
	podcast := createTestPodcast(500)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

// Benchmark feed filtering with different episode types
func BenchmarkGenerateRssFeed_WithFiltering(b *testing.B) {
	podcast := createTestPodcastWithMixedContent(100)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

// Benchmark with different podcast types
func BenchmarkGenerateRssFeed_ChannelType(b *testing.B) {
	podcast := createTestPodcast(50)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.CHANNEL, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

func BenchmarkGenerateRssFeed_PlaylistType(b *testing.B) {
	podcast := createTestPodcast(50)
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateRssFeed(podcast, host, enum.PLAYLIST, downloader.AudioFormat{Format: "m4a", Extension: ".m4a"})
	}
}

// Benchmark episode iteration and filtering
func BenchmarkFilterEpisodes(b *testing.B) {
	episodes := make([]models.PodcastEpisode, 100)
	for i := 0; i < 100; i++ {
		episodes[i] = models.PodcastEpisode{
			YoutubeVideoId:     "video" + string(rune(i)),
			EpisodeName:        "Episode " + string(rune(i)),
			EpisodeDescription: "Description " + string(rune(i)),
			Type:               "CHANNEL",
			Duration:           time.Duration(i+5) * time.Minute,
			PublishedDate:      time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var filtered []models.PodcastEpisode
		for _, ep := range episodes {
			if ep.Type == "CHANNEL" && ep.Duration.Seconds() >= 120 &&
				ep.EpisodeName != "Private video" &&
				ep.EpisodeDescription != "This video is private." {
				filtered = append(filtered, ep)
			}
		}
	}
}

// Helper functions for creating test data
func createTestPodcast(numEpisodes int) models.Podcast {
	episodes := make([]models.PodcastEpisode, numEpisodes)
	for i := 0; i < numEpisodes; i++ {
		episodes[i] = models.PodcastEpisode{
			YoutubeVideoId:     "video" + string(rune(i)),
			EpisodeName:        "Episode " + string(rune(i)),
			EpisodeDescription: "This is a test episode description for episode " + string(rune(i)),
			Type:               "CHANNEL",
			Duration:           10 * time.Minute,
			PublishedDate:      time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	return models.Podcast{
		Id:              "UC_test123",
		PodcastName:     "Test Podcast",
		Description:     "Test podcast description",
		ImageUrl:        "https://example.com/image.jpg",
		PodcastEpisodes: episodes,
		ArtistName:      "Test Artist",
		Category:        "Technology",
		Explicit:        "false",
		LastBuildDate:   time.Now().Format(time.RFC1123),
	}
}

func createTestPodcastWithMixedContent(numEpisodes int) models.Podcast {
	episodes := make([]models.PodcastEpisode, numEpisodes)
	for i := 0; i < numEpisodes; i++ {
		episode := models.PodcastEpisode{
			YoutubeVideoId: "video" + string(rune(i)),
			Type:           "CHANNEL",
			PublishedDate:  time.Now().Add(-time.Duration(i) * time.Hour),
		}

		// Create mixed content
		switch i % 5 {
		case 0:
			// Normal episode
			episode.EpisodeName = "Normal Episode " + string(rune(i))
			episode.EpisodeDescription = "Normal description"
			episode.Duration = 10 * time.Minute
		case 1:
			// Short episode (should be filtered)
			episode.EpisodeName = "Short Episode " + string(rune(i))
			episode.EpisodeDescription = "Short description"
			episode.Duration = 1 * time.Minute
		case 2:
			// Private video (should be filtered)
			episode.EpisodeName = "Private video"
			episode.EpisodeDescription = "This video is private."
			episode.Duration = 10 * time.Minute
		case 3:
			// Long episode
			episode.EpisodeName = "Long Episode " + string(rune(i))
			episode.EpisodeDescription = "Long description"
			episode.Duration = 60 * time.Minute
		case 4:
			// Playlist episode
			episode.EpisodeName = "Playlist Episode " + string(rune(i))
			episode.EpisodeDescription = "Playlist description"
			episode.Duration = 15 * time.Minute
			episode.Type = "PLAYLIST"
		}

		episodes[i] = episode
	}

	return models.Podcast{
		Id:              "UC_mixed_test",
		PodcastName:     "Mixed Content Podcast",
		Description:     "Podcast with mixed content types",
		ImageUrl:        "https://example.com/image.jpg",
		PodcastEpisodes: episodes,
		ArtistName:      "Test Artist",
		Category:        "Technology",
		Explicit:        "false",
		LastBuildDate:   time.Now().Format(time.RFC1123),
	}
}

package playlist

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/cache"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/webhook"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"
	"time"

	ytApi "google.golang.org/api/youtube/v3"

	"ikoyhn/podcast-sponsorblock/internal/logger"
)

func BuildPlaylistRssFeed(youtubePlaylistId string, params *models.RssRequestParams, host string, audioFormat downloader.AudioFormat) []byte {
	logger.Logger.Debug().Msg("[RSS FEED] Building rss feed for playlist...")

	// Check cache first
	rssFeedCache := cache.GetRSSFeedCache()
	cacheParams := map[string]interface{}{"host": host, "format": audioFormat.Format}
	if cachedFeed, found := rssFeedCache.Get("playlist", youtubePlaylistId, cacheParams); found {
		logger.Logger.Debug().Msg("[RSS FEED] Returning cached playlist feed")
		return cachedFeed
	}

	service := youtube.SetupYoutubeService()
	podcast := youtube.GetChannelData(youtubePlaylistId, service, true)

	getYoutubePlaylistData(youtubePlaylistId, service, params)
	episodes, err := database.GetPodcastEpisodesByPodcastId(youtubePlaylistId, enum.PLAYLIST)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return nil
	}

	// Apply filters and preferences
	episodes, err = database.ApplyFiltersAndPreferences(episodes, youtubePlaylistId)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to apply filters and preferences")
		// Continue without filters rather than failing completely
	}

	podcastRss := rss.BuildPodcast(podcast, episodes)
	rssFeed := rss.GenerateRssFeed(podcastRss, host, enum.PLAYLIST, audioFormat)

	// Store in cache
	rssFeedCache.Set("playlist", youtubePlaylistId, cacheParams, rssFeed)

	return rssFeed
}

func getYoutubePlaylistData(youtubePlaylistId string, service *ytApi.Service, params *models.RssRequestParams) {
	continueRequestingPlaylistItems := true
	var missingVideos []models.PodcastEpisode
	pageToken := "first_call"

	for continueRequestingPlaylistItems {
		// Create context with timeout for each API call
		ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)

		call := service.PlaylistItems.List([]string{"snippet", "status", "contentDetails"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(constants.PageSize).
			Context(ctx)
		call.Header().Set("order", "publishedAt desc")

		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()
		cancel() // Always cancel context after API call completes
		if ytAgainErr != nil {
			logger.Logger.Error().Msgf("Error calling YouTube API for Playlist: %s. Ensure your API key is valid, if your API key is valid you have have reached your API quota. Error: %v", youtubePlaylistId, ytAgainErr)
		}
		if response.HTTPStatusCode != http.StatusOK {
			logger.Logger.Error().Msgf("YouTube API returned status code %d for Playlist: %s", response.HTTPStatusCode, youtubePlaylistId)
			return
		}

		pageToken = response.NextPageToken
		for _, item := range response.Items {
			exists, err := database.EpisodeExists(item.Snippet.ResourceId.VideoId, "PLAYLIST")
			if err != nil {
				logger.Logger.Error().Err(err).Msg("")
			}
			if !exists {
				cleanedVideo := common.CleanPlaylistItems(item)
				if cleanedVideo != nil {
					missingVideos = append(missingVideos, models.NewPodcastEpisodeFromPlaylist(cleanedVideo))
				}
			} else {
				if len(missingVideos) > 0 {
					database.SavePlaylistEpisodes(missingVideos)
					// Send webhook notification for new episodes
					sendNewEpisodeWebhooks(missingVideos, youtubePlaylistId)
				}
				return
			}
		}
		if response.NextPageToken == "" {
			continueRequestingPlaylistItems = false
		}
	}
	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
		// Send webhook notification for new episodes
		sendNewEpisodeWebhooks(missingVideos, youtubePlaylistId)
	}
}

// sendNewEpisodeWebhooks sends webhook notifications for new episodes
func sendNewEpisodeWebhooks(episodes []models.PodcastEpisode, podcastId string) {
	podcast := database.GetPodcast(podcastId)
	podcastName := "Unknown"
	if podcast != nil {
		podcastName = podcast.PodcastName
	}

	for _, episode := range episodes {
		webhook.SendWebhook(webhook.EventNewEpisode, map[string]interface{}{
			"video_id":     episode.YoutubeVideoId,
			"title":        episode.EpisodeName,
			"podcast_id":   episode.PodcastId,
			"podcast_name": podcastName,
			"published_at": episode.PublishedDate.Format(time.RFC3339),
			"type":         episode.Type,
		})

		logger.Logger.Info().
			Str("video_id", episode.YoutubeVideoId).
			Str("title", episode.EpisodeName).
			Msg("Sent new episode webhook")
	}
}

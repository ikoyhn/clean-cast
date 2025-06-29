package playlist

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"

	ytApi "google.golang.org/api/youtube/v3"

	log "github.com/labstack/gommon/log"
)

func BuildPlaylistRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")

	service := youtube.SetupYoutubeService()
	podcast := youtube.GetChannelData(youtubePlaylistId, service, true)

	getYoutubePlaylistData(youtubePlaylistId, service)
	episodes, err := database.GetPodcastEpisodesByPodcastId(youtubePlaylistId)
	if err != nil {
		log.Error(err)
		return nil
	}

	podcastRss := rss.BuildPodcast(podcast, episodes)
	return rss.GenerateRssFeed(podcastRss, host, enum.PLAYLIST)
}

func getYoutubePlaylistData(youtubePlaylistId string, service *ytApi.Service) {
	log.Info("[RSS FEED] Getting youtube data...")
	continueRequestingPlaylistItems := true
	var missingVideos []models.PodcastEpisode
	pageToken := "first_call"

	for continueRequestingPlaylistItems {
		call := service.PlaylistItems.List([]string{"snippet", "status", "contentDetails"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(50)
		call.Header().Set("order", "publishedAt desc")

		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()
		if ytAgainErr != nil {
			log.Fatalf("Error calling YouTube API: %v. Ensure your API key is valid", response)
		}
		if response.HTTPStatusCode != http.StatusOK {
			log.Errorf("YouTube API returned status code %v", response.HTTPStatusCode)
			return
		}

		pageToken = response.NextPageToken
		for _, item := range response.Items {
			exists, err := database.EpisodeExists(item.Snippet.ResourceId.VideoId, "PLAYLIST")
			if err != nil {
				log.Error(err)
			}
			if !exists {
				cleanedVideo := common.CleanPlaylistItems(item)
				if cleanedVideo != nil {
					missingVideos = append(missingVideos, models.NewPodcastEpisodeFromPlaylist(cleanedVideo))
				}
			} else {
				if len(missingVideos) > 0 {
					database.SavePlaylistEpisodes(missingVideos)
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
	}
}

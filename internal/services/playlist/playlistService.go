package playlist

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"
	"time"

	log "github.com/labstack/gommon/log"
	ytApi "google.golang.org/api/youtube/v3"
)

func BuildPlaylistRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")

	dbPodcast := database.GetPodcast(youtubePlaylistId)

	shouldUpdate := true
	if dbPodcast != nil && dbPodcast.LastBuildDate != "" {
		dur, err := time.ParseDuration(config.AppConfig.Setup.PodcastRefreshInterval)
		if err != nil {
			panic("Invalid [podcast-refresh-interval] format. Use formats like '5m', '1h', '400s'.")
		}
		lastBuild, err := time.Parse(time.RFC1123, dbPodcast.LastBuildDate)
		if err == nil && time.Since(lastBuild) < dur {
			shouldUpdate = false
			log.Infof("[YOUTUBE API] Skipping channel update, last build date within %v", dur)
		}
	}

	if shouldUpdate {
		dbPodcast = youtube.GetChannelData(dbPodcast, youtubePlaylistId, false)
		getYoutubePlaylistData(youtubePlaylistId)
		dbPodcast = database.GetPodcast(youtubePlaylistId)
	}

	if dbPodcast == nil {
		dbPodcast = youtube.GetChannelData(dbPodcast, youtubePlaylistId, true)
	}

	episodes, err := database.GetPodcastEpisodesByPodcastId(youtubePlaylistId, enum.PLAYLIST)
	if err != nil {
		log.Error(err)
		return nil
	}

	podcastRss := rss.BuildPodcast(*dbPodcast, episodes)
	return rss.GenerateRssFeed(podcastRss, host, enum.PLAYLIST)
}

func getYoutubePlaylistData(youtubePlaylistId string) {
	continueRequestingPlaylistItems := true
	var missingVideos []models.PodcastEpisode
	pageToken := "first_call"
	isPlaylistDescOrder := true

	for continueRequestingPlaylistItems {
		call := youtube.YtService.PlaylistItems.List([]string{"snippet", "status", "contentDetails"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(50)
		call.Header().Set("order", "publishedAt desc")

		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()

		if ytAgainErr != nil {
			log.Errorf("Error calling YouTube API for Playlist: %w. Ensure your API key is valid, if your API key is valid you have have reached your API quota. Error: %w", youtubePlaylistId, response)
		}
		if response.HTTPStatusCode != http.StatusOK {
			log.Errorf("YouTube API returned status code %w for Playlist: %w", response.HTTPStatusCode, youtubePlaylistId)
			return
		}
		if pageToken == "first_call" {
			isPlaylistDescOrder = isPlaylistInDescOrder(response.Items)
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

				if isPlaylistDescOrder {
					return
				} else {
					log.Info("[YOUTUBE API] Playlist not in DESC order, grabbing all episodes")
				}
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

func isPlaylistInDescOrder(items []*ytApi.PlaylistItem) bool {
	if len(items) < 2 {
		return true
	}

	firstDate, _ := time.Parse(time.RFC3339, items[0].ContentDetails.VideoPublishedAt)
	lastDate, _ := time.Parse(time.RFC3339, items[len(items)-1].ContentDetails.VideoPublishedAt)

	return firstDate.After(lastDate)
}

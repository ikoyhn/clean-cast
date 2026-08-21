package playlist

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/http"
	"strconv"
	"time"

	log "github.com/labstack/gommon/log"
	ytApi "google.golang.org/api/youtube/v3"
)

func BuildPlaylistRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")
	dbPodcast := database.GetPodcast(youtubePlaylistId)

	shouldUpdate := true
	if dbPodcast != nil && dbPodcast.LastBuildDate != "" {
		dur := config.AppConfig.Setup.PodcastRefreshInterval
		lastBuild, err := common.ParseLastBuildDate(dbPodcast.LastBuildDate)
		if err != nil {
			log.Warnf("[YOUTUBE API] Unreadable last build date %q for playlist %s, refreshing: %v",
				dbPodcast.LastBuildDate, youtubePlaylistId, err)
		} else if time.Since(lastBuild) < dur {
			shouldUpdate = false
			log.Infof("[YOUTUBE API] Skipping playlist update, last build date within %v", dur)
		}
	}

	if shouldUpdate {
		updated, err := youtube.GetChannelData(dbPodcast, youtubePlaylistId, true)
		if err != nil {
			log.Errorf("[RSS FEED] Could not load playlist %s: %v", youtubePlaylistId, err)
			if dbPodcast == nil {
				return nil
			}
		} else {
			dbPodcast = updated
			getYoutubePlaylistData(youtubePlaylistId)
			if refreshed := database.GetPodcast(youtubePlaylistId); refreshed != nil {
				dbPodcast = refreshed
			}
		}
	}

	if dbPodcast == nil {
		log.Errorf("[RSS FEED] No podcast data available for playlist %s", youtubePlaylistId)
		return nil
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
	pageToken := "first_call"
	isPlaylistDescOrder := true

	for continueRequestingPlaylistItems {
		var missingVideoIds []string
		call := youtube.YtService.PlaylistItems.List([]string{"snippet", "status", "contentDetails"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(50)
		call.Header().Set("order", "publishedAt desc")

		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()
		if ytAgainErr != nil || response == nil {
			log.Errorf("Error calling YouTube API for Playlist: %s (%v). Ensure your API key is valid, if your API key is valid you have have reached your API quota.", youtubePlaylistId, ytAgainErr)
			return
		}
		if response.HTTPStatusCode != http.StatusOK {
			log.Errorf("YouTube API returned status code %s for Playlist: %s", strconv.Itoa(response.HTTPStatusCode), youtubePlaylistId)
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
					missingVideoIds = append(missingVideoIds, cleanedVideo.Snippet.ResourceId.VideoId)
				}
			} else {
				if isPlaylistDescOrder {
					continueRequestingPlaylistItems = false
					break
				} else {
					log.Info("[YOUTUBE API] Playlist not in DESC order, grabbing all episodes")
				}
			}
		}
		youtube.GetVideosAndValidate(missingVideoIds, enum.PLAYLIST, youtubePlaylistId)
		if response.NextPageToken == "" {
			continueRequestingPlaylistItems = false
		}
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

package services

import (
	"flag"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"math"
	"os"

	log "github.com/labstack/gommon/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
)

type Env struct {
	db *gorm.DB
}

var (
	maxResults = flag.Int64("max-results", 50, "Max YouTube results")
)

func BuildRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")

	ytData := getYoutubePlaylistData(youtubePlaylistId)
	allItems := cleanPlaylistItems(ytData)
	item := allItems[0]
	closestApplePodcastData := getAppleData(item, allItems)

	podcastRss := buildMainPodcast(ytData, closestApplePodcastData)
	return GenerateRssFeed(podcastRss, closestApplePodcastData, host)
}

func BuildChannelRssFeed(channelUsername string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for channel...")

	channelData, videoData, error := getChannelMetadataAndVideos(getChannelId(channelUsername))

	podcastRss := buildMainPodcast(ytData, closestApplePodcastData)
	return GenerateChannelRssFeed(podcastRss, closestApplePodcastData, host)
}

func buildMainPodcast(allItems []*youtube.PlaylistItem, appleData AppleResult) models.Podcast {
	item := allItems[0]
	return models.Podcast{
		AppleId:          appleData.CollectionId,
		YoutubePodcastId: item.Snippet.PlaylistId,
		PodcastName:      appleData.TrackName,
		Description:      appleData.TrackName,
		PostedDate:       appleData.ReleaseDate,
		ImageUrl: func() string {
			if appleData.ArtworkUrl100 != "" {
				return appleData.ArtworkUrl100
			}
			return getChannelProfilePicture(item.Snippet.ChannelId)
		}(),
		ArtistName:      appleData.ArtistName,
		Explicit:        appleData.ContentAdvisoryRating,
		PodcastEpisodes: buildPodcastEpisodes(allItems),
	}
}

func getAppleData(item *youtube.PlaylistItem, allItems []*youtube.PlaylistItem) AppleResult {
	itunesResponse := GetApplePodcastData(item.Snippet.ChannelTitle)
	closestApplePodcastData := findClosestResult(itunesResponse.Results, len(allItems))
	return closestApplePodcastData
}

func buildPodcastEpisodes(allItems []*youtube.PlaylistItem) []models.PodcastEpisode {
	podcastEpisodes := []models.PodcastEpisode{}
	for _, item := range allItems {
		tempPodcast := models.PodcastEpisode{
			YoutubeVideoId:     item.Snippet.ResourceId.VideoId,
			EpisodeName:        item.Snippet.Title,
			EpisodeDescription: item.Snippet.Description,
			Position:           item.Snippet.Position,
			PublishedDate:      item.Snippet.PublishedAt,
		}
		podcastEpisodes = append(podcastEpisodes, tempPodcast)

	}
	return podcastEpisodes
}

func DeterminePodcastDownload(youtubeVideoId string) (bool, float64) {
	episodeHistory := database.GetEpisodePlaybackHistory(youtubeVideoId)

	updatedSkippedTime := TotalSponsorTimeSkipped(youtubeVideoId)
	if episodeHistory == nil {
		return true, updatedSkippedTime
	}

	if math.Abs(episodeHistory.TotalTimeSkipped-updatedSkippedTime) > 2 {
		os.Remove("/config/audio/" + youtubeVideoId + ".m4a")
		log.Debug("[SponsorBlock] Updating downloaded episode with new sponsor skips...")
		return true, updatedSkippedTime
	}

	return false, updatedSkippedTime
}

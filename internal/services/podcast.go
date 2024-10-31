package services

import (
	"flag"
	log "github.com/labstack/gommon/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"ikoyhn/podcast-sponsorblock/internal/models"
)

var (
	maxResults = flag.Int64("max-results", 50, "Max YouTube results")
)

var UNAVAILABLE_STATUSES = []string{"private", "privacyStatusUnspecified"}

func BuildRssFeed(db *gorm.DB, youtubePlaylistId string, host string) []byte {
	log.Info("[RSS FEED] Building rss feed...")

	ytData := getYoutubeData(youtubePlaylistId)
	allItems := cleanPlaylistItems(ytData)
	item := allItems[0]
	closestApplePodcastData := getAppleData(item, allItems)

	podcastRss := buildMainPodcast(ytData, closestApplePodcastData)
	return GenerateRssFeed(podcastRss, closestApplePodcastData, host)
}

func buildMainPodcast(allItems []*youtube.PlaylistItem, appleData AppleResult) models.Podcast {
	item := allItems[0]
	return models.Podcast{
		AppleId:          appleData.CollectionId,
		YoutubePodcastId: item.Snippet.PlaylistId,
		PodcastName:      appleData.TrackName,
		Description:      appleData.TrackName,
		PostedDate:       appleData.ReleaseDate,
		ImageUrl:         appleData.ArtworkUrl100,
		ArtistName:       appleData.ArtistName,
		Explicit:         appleData.ContentAdvisoryRating,
		PodcastEpisodes:  buildPodcastEpisodes(allItems),
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

package services

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"

	log "github.com/labstack/gommon/log"
)

func BuildPlaylistRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")

	service := setupYoutubeService()
	podcast := getChannelData(youtubePlaylistId, service, true)

	getYoutubePlaylistData(youtubePlaylistId, service)
	episodes, err := database.GetPodcastEpisodesByPodcastId(youtubePlaylistId)
	if err != nil {
		log.Error(err)
		return nil
	}

	podcastRss := buildPodcast(podcast, episodes)
	return GenerateRssFeed(podcastRss, host)
}

func buildPodcast(podcast models.Podcast, allItems []models.PodcastEpisode) models.Podcast {
	podcast.PodcastEpisodes = allItems
	return podcast
}

package services

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"math"
	"os"

	log "github.com/labstack/gommon/log"
)

func BuildChannelRssFeed(channelId string, host string) []byte {
	log.Info("[RSS FEED] Building rss feed for channel...")
	service := setupYoutubeService()

	podcast := getChannelData(channelId, service, false)

	getChannelMetadataAndVideos(podcast.Id, service)
	episodes, err := database.GetPodcastEpisodesByPodcastId(podcast.Id)
	if err != nil {
		log.Error(err)
		return nil
	}

	podcastRss := buildPodcast(podcast, episodes)
	return GenerateRssFeed(podcastRss, host)
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

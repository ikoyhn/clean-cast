package services

import (
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"math"
	"os"
	"strconv"

	log "github.com/labstack/gommon/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
)

type Env struct {
	db *gorm.DB
}

func BuildRssFeed(youtubePlaylistId string, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for playlist...")

	ytData := getYoutubePlaylistData(youtubePlaylistId)
	allItems := cleanPlaylistItems(ytData)
	item := allItems[0]
	closestApplePodcastData := getAppleData(item.Snippet.ChannelTitle, len(allItems))

	podcastRss := buildMainPlaylistPodcast(ytData, closestApplePodcastData)
	return GenerateRssFeed(podcastRss, closestApplePodcastData, host)
}

func BuildChannelRssFeed(channelId string, host string) []byte {
	log.Info("[RSS FEED] Building rss feed for channel...")

	channelData, videoData, error := getChannelMetadataAndVideos(channelId)
	if error != nil {
		log.Error(error)
	}

	closestApplePodcastData := getAppleData(channelData.Title, len(videoData))
	podcastRss := buildMainChannelPodcast(channelData, videoData, closestApplePodcastData)
	return GenerateRssFeed(podcastRss, closestApplePodcastData, host)
}

func buildMainChannelPodcast(metaData ChannelMetadata, videos []VideoMetadata, appleData AppleResult) models.Podcast {
	appleId := ""
	podcastName := ""
	description := ""
	postedDate := ""
	imageUrl := ""
	artistName := ""
	explicit := ""

	if (appleData == AppleResult{}) {
		log.Info("[RSS FEED] No Apple data found")
		appleId = metaData.Id
		podcastName = metaData.Title
		description = metaData.Description
		postedDate = ""
		imageUrl = metaData.ThumbnailURL
		artistName = metaData.Title
		explicit = ""
		log.Info("[RSS FEED] Image url: ", imageUrl)
	} else {
		appleId = strconv.Itoa(appleData.CollectionId)
		podcastName = appleData.TrackName
		description = appleData.TrackName
		postedDate = appleData.ReleaseDate
		imageUrl = appleData.ArtworkUrl100
		artistName = appleData.ArtistName
		explicit = appleData.ContentAdvisoryRating
	}

	return models.Podcast{
		AppleId:          appleId,
		YoutubePodcastId: metaData.Id,
		PodcastName:      podcastName,
		Description:      description,
		PostedDate:       postedDate,
		ImageUrl:         imageUrl,
		ArtistName:       artistName,
		Explicit:         explicit,
		PodcastEpisodes:  buildChannelPodcastEpisodes(videos),
	}
}

func buildMainPlaylistPodcast(allItems []*youtube.PlaylistItem, appleData AppleResult) models.Podcast {
	item := allItems[0]
	log.Info("[RSS FEED] apple url: ", item.Snippet.ChannelTitle)
	log.Info("[RSS FEED] Channel profile picture: ", getChannelProfilePicture(item.Snippet.ChannelId))
	return models.Podcast{
		AppleId:          strconv.Itoa(appleData.CollectionId),
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
		PodcastEpisodes: buildPlaylistPodcastEpisodes(allItems),
	}
}

func getAppleData(channelTitle string, numOfVideos int) AppleResult {
	itunesResponse := GetApplePodcastData(channelTitle)
	closestApplePodcastData := findClosestResult(itunesResponse.Results, numOfVideos)
	return closestApplePodcastData
}

func buildChannelPodcastEpisodes(allItems []VideoMetadata) []models.PodcastEpisode {
	podcastEpisodes := []models.PodcastEpisode{}
	position := int64(0)
	for _, item := range allItems {
		tempPodcast := models.PodcastEpisode{
			YoutubeVideoId:     item.VideoID,
			EpisodeName:        item.Title,
			EpisodeDescription: item.Description,
			Position:           position,
			PublishedDate:      item.PublishedAt,
		}
		podcastEpisodes = append(podcastEpisodes, tempPodcast)
		position++
	}
	return podcastEpisodes
}

func buildPlaylistPodcastEpisodes(allItems []*youtube.PlaylistItem) []models.PodcastEpisode {
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

package channel

import (
	"context"
	"errors"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/webhook"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"time"

	"gorm.io/gorm"

	"ikoyhn/podcast-sponsorblock/internal/logger"

	ytApi "google.golang.org/api/youtube/v3"
)

func GetChannelMetadataAndVideos(channelId string, service *ytApi.Service, params *models.RssRequestParams) {
	logger.Logger.Info().Msg("[RSS FEED] Getting channel data...")

	if !youtube.FindChannel(channelId, service) {
		return
	}
	oldestSavedEpisode, err := database.GetOldestEpisode(channelId)
	latestSavedEpisode, err := database.GetLatestEpisode(channelId)

	switch determineRequestType(params) {
	case enum.DATE:
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		if oldestSavedEpisode != nil {
			if latestSavedEpisode.PublishedDate.After(*params.Date) {
				getChannelVideosByDateRange(service, channelId, time.Now(), *params.Date)
			} else if oldestSavedEpisode.PublishedDate.After(*params.Date) {
				getChannelVideosByDateRange(service, channelId, oldestSavedEpisode.PublishedDate, *params.Date)
				getChannelVideosByDateRange(service, channelId, time.Now(), latestSavedEpisode.PublishedDate)
			}
		} else {
			getChannelVideosByDateRange(service, channelId, time.Now(), *params.Date)
		}
	case enum.DEFAULT:
		if (oldestSavedEpisode != nil) && (latestSavedEpisode != nil) {
			getChannelVideosByDateRange(service, channelId, time.Now(), latestSavedEpisode.PublishedDate)
			getChannelVideosByDateRange(service, channelId, oldestSavedEpisode.PublishedDate, time.Unix(0, 0))
		} else {
			getChannelVideosByDateRange(service, channelId, time.Unix(0, 0), time.Unix(0, 0))
		}
	}
}

func getChannelVideosByDateRange(service *ytApi.Service, channelID string, beforeDateParam time.Time, afterDateParam time.Time) {

	savedEpisodeIds, err := database.GetAllPodcastEpisodeIds(channelID)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return
	}

	nextPageToken := ""
	for {
		var videoIdsNotSaved []string

		// Create context with timeout for each API call
		ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)

		searchCall := service.Search.List([]string{"id", "snippet"}).
			ChannelId(channelID).
			Type("video").
			Order("date").
			MaxResults(constants.PageSize).
			PageToken(nextPageToken).
			Context(ctx)
		if !afterDateParam.IsZero() {
			searchCall = searchCall.PublishedAfter(afterDateParam.Format(time.RFC3339))
		}
		if !beforeDateParam.IsZero() {
			searchCall = searchCall.PublishedBefore(beforeDateParam.Format(time.RFC3339))
		}
		searchCallResponse, err := searchCall.Do()
		cancel() // Always cancel context after API call completes
		if err != nil {
			logger.Logger.Error().Err(err).Msg("")
			return
		}

		videoIdsNotSaved = append(videoIdsNotSaved, getValidVideosFromChannelResponse(searchCallResponse, savedEpisodeIds)...)
		if len(videoIdsNotSaved) > 0 {
			fetchAndSaveVideos(service, videoIdsNotSaved)
		}

		nextPageToken = searchCallResponse.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
}

func getValidVideosFromChannelResponse(channelVideoResponse *ytApi.SearchListResponse, savedEpisodeIds []string) []string {
	var videoIds []string
	var filteredItems []*ytApi.SearchResult
	for _, item := range channelVideoResponse.Items {
		if !common.Contains(savedEpisodeIds, item.Id.VideoId) && (item.Id.Kind == "youtube#video" || item.Id.Kind == "youtube#searchResult") {
			filteredItems = append(filteredItems, item)
			videoIds = append(videoIds, item.Id.VideoId)
		}
	}
	channelVideoResponse.Items = filteredItems
	return videoIds
}

func fetchAndSaveVideos(service *ytApi.Service, videoIdsNotSaved []string) {
	var missingVideos []models.PodcastEpisode
	missingVideos, err := youtube.GetVideoAndValidate(service, videoIdsNotSaved, missingVideos)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return
	}

	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
		// Send webhook notification for new episodes
		sendNewEpisodeWebhooks(missingVideos)
	}
}

// sendNewEpisodeWebhooks sends webhook notifications for new episodes
func sendNewEpisodeWebhooks(episodes []models.PodcastEpisode) {
	for _, episode := range episodes {
		podcast := database.GetPodcast(episode.PodcastId)
		podcastName := "Unknown"
		if podcast != nil {
			podcastName = podcast.PodcastName
		}

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

func determineRequestType(params *models.RssRequestParams) enum.PodcastFetchType {
	if params.Date != nil {
		return enum.DATE
	} else {
		return enum.DEFAULT
	}
}

package channel

import (
	"errors"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"time"

	"gorm.io/gorm"

	log "github.com/labstack/gommon/log"

	ytApi "google.golang.org/api/youtube/v3"
)

func GetChannelMetadataAndVideos(channelId string, service *ytApi.Service, params *models.RssRequestParams) {
	log.Info("[RSS FEED] Getting channel data...")

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
			getChannelVideosByDateRange(service, channelId, time.Now(), oldestSavedEpisode.PublishedDate)
			if oldestSavedEpisode.PublishedDate.After(*params.Date) {
				getChannelVideosByDateRange(service, channelId, oldestSavedEpisode.PublishedDate, *params.Date)
			}
		} else {
			getChannelVideosByDateRange(service, channelId, time.Now(), *params.Date)
		}
	case enum.DEFAULT:
		if (oldestSavedEpisode != nil) && (latestSavedEpisode != nil) {
			getChannelVideosByDateRange(service, channelId, time.Now(), latestSavedEpisode.PublishedDate)
			getChannelVideosByDateRange(service, channelId, oldestSavedEpisode.PublishedDate, time.Unix(0, 0))
		} else {
			getChannelVideosByDateRange(service, channelId, time.Now(), time.Unix(0, 0))
		}
	}
}

func getChannelVideosByDateRange(service *ytApi.Service, channelID string, beforeDateParam time.Time, afterDateParam time.Time) {
	var pageToken string
	var videoIdsNotSaved []string
	savedEpisodeIds, err := database.GetAllPodcastEpisodeIds(channelID)
	if err != nil {
		log.Error(err)
		return
	}

	for {
		channelVideosCall := service.Search.List([]string{"snippet"}).
			ChannelId(channelID).
			PublishedAfter(afterDateParam.Format(time.RFC3339)).
			PublishedBefore(beforeDateParam.Format(time.RFC3339)).
			Order("date")
		if pageToken != "" {
			channelVideosCall.PageToken(pageToken)
		}
		channelVideoResponse, err := channelVideosCall.Do()
		if err != nil {
			log.Error(err)
			return
		}

		videoIdsNotSaved = append(videoIdsNotSaved, getValidVideosFromChannelResponse(channelVideoResponse, savedEpisodeIds)...)
		fetchAndSaveVideos(service, videoIdsNotSaved, channelVideoResponse, &pageToken)

		if channelVideoResponse.NextPageToken != "" {
			pageToken = channelVideoResponse.NextPageToken
		} else {
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

func fetchAndSaveVideos(service *ytApi.Service, videoIdsNotSaved []string, channelVideoResponse *ytApi.SearchListResponse, pageToken *string) {
	var missingVideos []models.PodcastEpisode
	missingVideos = youtube.GetVideoAndValidate(service, videoIdsNotSaved, missingVideos)

	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
	}

	*pageToken = channelVideoResponse.NextPageToken
}

func determineRequestType(params *models.RssRequestParams) enum.PodcastFetchType {
	if params.Date != nil {
		return enum.DATE
	} else {
		return enum.DEFAULT
	}
}

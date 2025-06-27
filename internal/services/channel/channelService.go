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
		log.Error(err)
		return
	}

	nextPageToken := ""
	for {
		var videoIdsNotSaved []string
		searchCall := service.Search.List([]string{"id", "snippet"}).
			ChannelId(channelID).
			Type("video").
			Order("date").
			MaxResults(50).
			PageToken(nextPageToken)
		if !afterDateParam.IsZero() {
			searchCall = searchCall.PublishedAfter(afterDateParam.Format(time.RFC3339))
		}
		if !beforeDateParam.IsZero() {
			searchCall = searchCall.PublishedBefore(beforeDateParam.Format(time.RFC3339))
		}
		searchCallResponse, err := searchCall.Do()
		if err != nil {
			log.Error(err)
			return
		}

		videoIdsNotSaved = append(videoIdsNotSaved, getValidVideosFromChannelResponse(searchCallResponse, savedEpisodeIds)...)
		if videoIdsNotSaved != nil && len(videoIdsNotSaved) > 0 {
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
	missingVideos = youtube.GetVideoAndValidate(service, videoIdsNotSaved, missingVideos)

	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
	}
}

func determineRequestType(params *models.RssRequestParams) enum.PodcastFetchType {
	if params.Date != nil {
		return enum.DATE
	} else {
		return enum.DEFAULT
	}
}

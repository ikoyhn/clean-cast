package channel

import (
	"errors"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"ikoyhn/podcast-sponsorblock/internal/services/rss"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"time"

	"gorm.io/gorm"

	log "github.com/labstack/gommon/log"

	ytApi "google.golang.org/api/youtube/v3"
)

func BuildChannelRssFeed(channelId string, params *models.RssRequestParams, host string) []byte {
	log.Debug("[RSS FEED] Building rss feed for channel...")
	dbPodcast := database.GetPodcast(channelId)

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
		updated, err := youtube.GetChannelData(dbPodcast, channelId, false)
		if err != nil {
			log.Errorf("[RSS FEED] Could not load channel %s: %v", channelId, err)
			if dbPodcast == nil {
				return nil
			}
		} else {
			dbPodcast = updated
			getChannelMetadataAndVideos(channelId, params)
			if refreshed := database.GetPodcast(channelId); refreshed != nil {
				dbPodcast = refreshed
			}
		}
	}

	if dbPodcast == nil {
		log.Errorf("[RSS FEED] No podcast data available for channel %s", channelId)
		return nil
	}

	episodes, err := database.GetPodcastEpisodesByPodcastId(channelId, enum.CHANNEL)
	if err != nil {
		log.Error(err)
		return nil
	}

	podcastRss := rss.BuildPodcast(*dbPodcast, episodes)
	return rss.GenerateRssFeed(podcastRss, host, enum.CHANNEL)
}

func getChannelMetadataAndVideos(channelId string, params *models.RssRequestParams) {
	log.Info("[RSS FEED] Getting channel data...")

	if !youtube.FindChannel(channelId) {
		return
	}
	oldestSavedEpisode, oldestErr := database.GetOldestEpisode(channelId)
	if oldestErr != nil && !errors.Is(oldestErr, gorm.ErrRecordNotFound) {
		log.Error(oldestErr)
		return
	}
	latestSavedEpisode, latestErr := database.GetLatestEpisode(channelId)
	if latestErr != nil && !errors.Is(latestErr, gorm.ErrRecordNotFound) {
		log.Error(latestErr)
		return
	}

	switch determineRequestType(params) {
	case enum.DATE:
		if oldestSavedEpisode != nil && latestSavedEpisode != nil {
			if latestSavedEpisode.PublishedDate.After(*params.Date) {
				getChannelVideosByDateRange(channelId, time.Now(), *params.Date)
			} else if oldestSavedEpisode.PublishedDate.After(*params.Date) {
				getChannelVideosByDateRange(channelId, oldestSavedEpisode.PublishedDate, *params.Date)
				getChannelVideosByDateRange(channelId, time.Now(), latestSavedEpisode.PublishedDate)
			}
		} else {
			getChannelVideosByDateRange(channelId, time.Now(), *params.Date)
		}
	case enum.DEFAULT:
		if (oldestSavedEpisode != nil) && (latestSavedEpisode != nil) {
			getChannelVideosByDateRange(channelId, time.Now(), latestSavedEpisode.PublishedDate)
			getChannelVideosByDateRange(channelId, oldestSavedEpisode.PublishedDate, time.Unix(0, 0))
		} else {
			getChannelVideosByDateRange(channelId, time.Unix(0, 0), time.Unix(0, 0))
		}
	}
}

func getChannelVideosByDateRange(channelID string, beforeDateParam time.Time, afterDateParam time.Time) {

	savedEpisodeIds, err := database.GetAllPodcastEpisodeIds(channelID)
	if err != nil {
		log.Error(err)
		return
	}

	nextPageToken := ""
	for {
		var videoIdsNotSaved []string
		searchCall := youtube.YtService.Search.List([]string{"id", "snippet"}).
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
		if len(videoIdsNotSaved) > 0 {
			youtube.GetVideosAndValidate(videoIdsNotSaved, enum.CHANNEL, channelID)
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

func determineRequestType(params *models.RssRequestParams) enum.PodcastFetchType {
	if params == nil || params.Date == nil {
		return enum.DEFAULT
	} else {
		return enum.DATE
	}
}

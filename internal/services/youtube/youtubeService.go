package youtube

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"time"

	"google.golang.org/api/option"
	ytApi "google.golang.org/api/youtube/v3"
	"ikoyhn/podcast-sponsorblock/internal/logger"
)

func GetChannelData(channelIdentifier string, service *ytApi.Service, isPlaylist bool) models.Podcast {
	var channelCall *ytApi.ChannelsListCall
	var channelId string
	dbPodcast := database.GetPodcast(channelIdentifier)

	if dbPodcast == nil {
		// Create context with timeout for API calls
		ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)
		defer cancel()

		if isPlaylist {
			playlistCall := service.Playlists.List([]string{"snippet", "status", "contentDetails"}).
				Id(channelIdentifier).
				Context(ctx)
			playlistResponse, err := playlistCall.Do()
			if err != nil {
				logger.Logger.Error().Msgf("Error retrieving playlist details: %v", err)
			}
			if len(playlistResponse.Items) == 0 {
				logger.Logger.Error().Msgf("Playlist not found")
			}
			playlist := playlistResponse.Items[0]
			channelId = playlist.Snippet.ChannelId
		} else {
			channelId = channelIdentifier
		}

		channelCall = service.Channels.List([]string{"snippet", "statistics", "contentDetails"}).
			Id(channelId).
			Context(ctx)
		channelResponse, err := channelCall.Do()
		if err != nil {
			logger.Logger.Error().Msgf("Error retrieving channel details: %v", err)
		}
		if len(channelResponse.Items) == 0 {
			logger.Logger.Error().Msgf("Channel not found")
		}
		channel := channelResponse.Items[0]

		// Use common utility for thumbnail selection
		imageUrl := common.SelectBestThumbnail(channel.Snippet.Thumbnails)

		dbPodcast = &models.Podcast{
			Id:              channelIdentifier,
			PodcastName:     channel.Snippet.Title,
			Description:     channel.Snippet.Description,
			ImageUrl:        imageUrl,
			PostedDate:      channel.Snippet.PublishedAt,
			PodcastEpisodes: []models.PodcastEpisode{},
			ArtistName:      channel.Snippet.Title,
			Explicit:        "false",
		}

		dbPodcast.LastBuildDate = time.Now().Format(time.RFC1123)
		database.SavePodcast(dbPodcast)
	}

	return *dbPodcast
}

func GetVideoAndValidate(service *ytApi.Service, videoIdsNotSaved []string, missingVideos []models.PodcastEpisode) ([]models.PodcastEpisode, error) {
	// Create context with timeout for API calls
	ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)
	defer cancel()

	videoCall := service.Videos.List([]string{"id,snippet,contentDetails"}).
		Id(videoIdsNotSaved...).
		MaxResults(int64(len(videoIdsNotSaved))).
		Context(ctx)

	videoResponse, err := videoCall.Do()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return nil, err
	}

	dur, err := time.ParseDuration(config.Config.MinDuration)
	if err != nil {
		logger.Logger.Error().Msgf("Invalid MIN_DURATION format '%s'. Use formats like '5m', '1h', '400s'. Error: %v", config.Config.MinDuration, err)
		return nil, err
	}

	for _, item := range videoResponse.Items {
		if item.Id != "" {
			duration, err := common.ParseDuration(item.ContentDetails.Duration)
			if err != nil {
				logger.Logger.Error().Err(err).Msg("")
				continue
			}

			if duration.Seconds() > dur.Seconds() {
				if database.IsEpisodeSaved(item) {
					return missingVideos, nil
				}
				missingVideos = append(missingVideos, models.NewPodcastEpisodeFromSearch(item, duration))
			}
		}
	}
	return missingVideos, nil
}

func FindChannel(channelID string, service *ytApi.Service) bool {
	exists, err := database.PodcastExists(channelID)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return true
	}

	if !exists {
		// Create context with timeout for API calls
		ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)
		defer cancel()

		channelCall := service.Channels.List([]string{"snippet", "statistics", "contentDetails"}).
			Id(channelID).
			Context(ctx)

		channelResponse, err := channelCall.Do()
		if err != nil {
			logger.Logger.Error().Err(err).Msg("")
			return false
		}

		if len(channelResponse.Items) == 0 {
			logger.Logger.Error("channel not found")
			return false
		}
	}
	return true
}

func SetupYoutubeService() *ytApi.Service {
	apiKey := config.Config.GoogleApiKey
	// Use context with timeout for YouTube API setup
	ctx, cancel := context.WithTimeout(context.Background(), constants.RequestTimeout*time.Second)
	defer cancel()

	service, err := ytApi.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		logger.Logger.Error().Msgf("Error creating new YouTube client: %v", err)
	}
	if service == nil {
		logger.Logger.Error().Msgf("Failed to create YouTube service: %v", err)
	}
	return service
}

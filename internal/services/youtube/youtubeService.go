package youtube

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"time"

	log "github.com/labstack/gommon/log"
	"google.golang.org/api/option"
	ytApi "google.golang.org/api/youtube/v3"
)

var YtService *ytApi.Service

func SetupYoutubeService() {
	apiKey := config.AppConfig.Setup.GoogleApiKey
	ctx := context.Background()
	service, err := ytApi.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil || service == nil {
		panic(fmt.Sprintf("Failed to create YouTube service: %v", err))
	}
	YtService = service
}
func GetChannelData(dbPodcast *models.Podcast, channelIdentifier string, isPlaylist bool) (*models.Podcast, error) {
	var channelCall *ytApi.ChannelsListCall
	var channelId string

	if dbPodcast == nil {
		if isPlaylist {
			playlistCall := YtService.Playlists.List([]string{"snippet", "status", "contentDetails"}).
				Id(channelIdentifier)
			playlistResponse, err := playlistCall.Do()
			if err != nil {
				return nil, fmt.Errorf("retrieving playlist details for %s: %w", channelIdentifier, err)
			}
			if len(playlistResponse.Items) == 0 {
				return nil, fmt.Errorf("playlist not found: %s", channelIdentifier)
			}
			playlist := playlistResponse.Items[0]
			channelId = playlist.Snippet.ChannelId
		} else {
			channelId = channelIdentifier
		}

		channelCall = YtService.Channels.List([]string{"snippet", "statistics", "contentDetails"}).
			Id(channelId)
		channelResponse, err := channelCall.Do()
		if err != nil {
			return nil, fmt.Errorf("retrieving channel details for %s: %w", channelId, err)
		}
		if len(channelResponse.Items) == 0 {
			return nil, fmt.Errorf("channel not found: %s", channelId)
		}
		channel := channelResponse.Items[0]

		imageUrl := ""
		if channel.Snippet.Thumbnails.Maxres != nil {
			imageUrl = channel.Snippet.Thumbnails.Maxres.Url
		} else if channel.Snippet.Thumbnails.Standard != nil {
			imageUrl = channel.Snippet.Thumbnails.Standard.Url
		} else if channel.Snippet.Thumbnails.High != nil {
			imageUrl = channel.Snippet.Thumbnails.High.Url
		} else if channel.Snippet.Thumbnails.Default != nil {
			imageUrl = channel.Snippet.Thumbnails.Default.Url
		}

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
	}
	dbPodcast.LastBuildDate = common.FormatLastBuildDate(time.Now())
	database.UpdatePodcast(dbPodcast)

	return dbPodcast, nil
}

func GetVideosAndValidate(videoIdsNotSaved []string, podcastType enum.PodcastType, podcastId string) {
	if len(videoIdsNotSaved) == 0 {
		return
	}
	var missingVideos []models.PodcastEpisode
	videoCall := YtService.Videos.List([]string{"id,snippet,contentDetails"}).
		Id(videoIdsNotSaved...).
		MaxResults(int64(len(videoIdsNotSaved)))

	videoResponse, err := videoCall.Do()
	if err != nil {
		log.Errorf("Error retrieving video details: %v", err)
		return
	}

	dur := config.AppConfig.Ytdlp.EpisodeDurationMinimum

	for _, item := range videoResponse.Items {
		if item.Id != "" {
			duration, err := common.ParseDuration(item.ContentDetails.Duration)
			if err != nil {
				log.Error(err)
				continue
			}

			if duration.Seconds() > dur.Seconds() {
				if database.IsEpisodeSaved(item) {
					continue
				}
				missingVideos = append(missingVideos, models.NewPodcastEpisode(item, duration, podcastType, podcastId))
			}
		}
	}
	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
	}
}

func FindChannel(channelID string) bool {
	exists, err := database.PodcastExists(channelID)
	if err != nil {
		log.Error(err)
		return true
	}

	if !exists {
		channelCall := YtService.Channels.List([]string{"snippet", "statistics", "contentDetails"})
		channelCall = channelCall.Id(channelID)

		channelResponse, err := channelCall.Do()
		if err != nil {
			log.Error(err)
			return false
		}

		if len(channelResponse.Items) == 0 {
			log.Error("channel not found")
			return false
		}
	}
	return true
}

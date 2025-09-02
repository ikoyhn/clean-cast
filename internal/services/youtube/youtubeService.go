package youtube

import (
	"context"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/common"
	"time"

	log "github.com/labstack/gommon/log"
	"google.golang.org/api/option"
	ytApi "google.golang.org/api/youtube/v3"
)

func GetChannelData(channelIdentifier string, service *ytApi.Service, isPlaylist bool) models.Podcast {
	var channelCall *ytApi.ChannelsListCall
	var channelId string
	dbPodcast := database.GetPodcast(channelIdentifier)

	if dbPodcast == nil {
		if isPlaylist {
			playlistCall := service.Playlists.List([]string{"snippet", "status", "contentDetails"}).
				Id(channelIdentifier)
			playlistResponse, err := playlistCall.Do()
			if err != nil {
				log.Errorf("Error retrieving playlist details: %v", err)
			}
			if len(playlistResponse.Items) == 0 {
				log.Errorf("Playlist not found")
			}
			playlist := playlistResponse.Items[0]
			channelId = playlist.Snippet.ChannelId
		} else {
			channelId = channelIdentifier
		}

		channelCall = service.Channels.List([]string{"snippet", "statistics", "contentDetails"}).
			Id(channelId)
		channelResponse, err := channelCall.Do()
		if err != nil {
			log.Errorf("Error retrieving channel details: %v", err)
		}
		if len(channelResponse.Items) == 0 {
			log.Errorf("Channel not found")
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

		dbPodcast.LastBuildDate = time.Now().Format(time.RFC1123)
		database.SavePodcast(dbPodcast)
	}

	return *dbPodcast
}

func GetVideoAndValidate(service *ytApi.Service, videoIdsNotSaved []string, missingVideos []models.PodcastEpisode) []models.PodcastEpisode {
	videoCall := service.Videos.List([]string{"id,snippet,contentDetails"}).
		Id(videoIdsNotSaved...).
		MaxResults(int64(len(videoIdsNotSaved)))

	videoResponse, err := videoCall.Do()
	if err != nil {
		log.Error(err)
		return nil
	}

	for _, item := range videoResponse.Items {
		if item.Id != "" {
			duration, err := common.ParseDuration(item.ContentDetails.Duration)
			if err != nil {
				log.Error(err)
				continue
			}
			if duration.Seconds() > 1000 {
				if database.IsEpisodeSaved(item) {
					return missingVideos
				}
				missingVideos = append(missingVideos, models.NewPodcastEpisodeFromSearch(item, duration))
			}
		}
	}
	return missingVideos
}

func FindChannel(channelID string, service *ytApi.Service) bool {
	exists, err := database.PodcastExists(channelID)
	if err != nil {
		log.Error(err)
		return true
	}

	if !exists {
		channelCall := service.Channels.List([]string{"snippet", "statistics", "contentDetails"})
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

func SetupYoutubeService() *ytApi.Service {
	apiKey := config.Config.GoogleApiKey
	ctx := context.Background()
	service, err := ytApi.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Errorf("Error creating new YouTube client: %v", err)
	}
	if service == nil {
		log.Errorf("Failed to create YouTube service: %v", err)
	}
	return service
}

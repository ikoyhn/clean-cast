package services

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const youtubeVideoUrl = "https://www.youtube.com/watch?v="

var youtubeVideoMutexes = &sync.Map{}

// Get all youtube playlist items and meta data for the RSS feed
func getYoutubePlaylistData(youtubePlaylistId string, service *youtube.Service) {

	log.Info("[RSS FEED] Getting youtube data...")

	continue_requesting_playlist_items := true
	missingVideos := []models.PodcastEpisode{}
	pageToken := "first_call"
	for continue_requesting_playlist_items {
		call := service.PlaylistItems.List([]string{"snippet", "status"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(50)
		call.Header().Set("order", "publishedAt desc")

		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()
		if ytAgainErr != nil {
			log.Fatalf("Error calling YouTube API: %v. Ensure your API key is valid", response)
		}
		if response.HTTPStatusCode != http.StatusOK {
			log.Errorf("YouTube API returned status code %v", response.HTTPStatusCode)
			return
		}

		pageToken = response.NextPageToken
		for _, item := range response.Items {
			exists, err := database.EpisodeExists(item.Snippet.ResourceId.VideoId, "PLAYLIST")
			if err != nil {
				log.Error(err)
			}
			if !exists {
				cleanedVideo := cleanPlaylistItems(item)
				if cleanedVideo != nil {
					missingVideos = append(missingVideos, models.NewPodcastEpisodeFromPlaylist(cleanedVideo))
				}
			} else {
				if len(missingVideos) > 0 {
					database.SavePlaylistEpisodes(missingVideos)
				}
				return
			}
		}
		if response.NextPageToken == "" {
			continue_requesting_playlist_items = false
		}
	}
	if len(missingVideos) > 0 {
		database.SavePlaylistEpisodes(missingVideos)
	}
}

func getChannelData(channelIdentifier string, service *youtube.Service, isPlaylist bool) models.Podcast {
	var channelCall *youtube.ChannelsListCall
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
			Id:              channel.Id,
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
func getChannelMetadataAndVideos(channelID string, service *youtube.Service) {
	log.Info("[RSS FEED] Getting channel data...")
	doesChannelExist := findChannel(channelID, service)
	if !doesChannelExist {
		return
	}

	missingVideos := []models.PodcastEpisode{}
	nextPageToken := ""
	for {
		channelVideosCall := service.Search.List([]string{"snippet"})
		channelVideosCall.ChannelId(channelID)
		channelVideosCall.Order("date")
		channelVideosCall.MaxResults(50)
		channelVideosCall.PageToken(nextPageToken)
		channelVideosCall.Type("video")

		channelVideoResponse, err := channelVideosCall.Do()
		if err != nil {
			log.Error(err)
			return
		}

		videoIds := getUnSavedEpisodeIds(channelVideoResponse)
		missingVideos = findMissingPodcastEpisodes(service, videoIds, missingVideos)

		if len(missingVideos) > 0 {
			database.SavePlaylistEpisodes(missingVideos)
			missingVideos = []models.PodcastEpisode{}
		} else {
			break
		}

		if channelVideoResponse.NextPageToken == "" {
			break
		}
		nextPageToken = channelVideoResponse.NextPageToken
	}
}

func findChannel(channelID string, service *youtube.Service) bool {
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

func getUnSavedEpisodeIds(channelVideoResponse *youtube.SearchListResponse) []string {
	videoIds := []string{}
	for _, item := range channelVideoResponse.Items {
		if item.Id.Kind == "youtube#video" {
			exists, err := database.EpisodeExists(item.Id.VideoId, "CHANNEL")
			if err != nil {
				log.Error(err)
			}
			if !exists {
				videoIds = append(videoIds, item.Id.VideoId)
			}
		}
	}
	return videoIds
}

func findMissingPodcastEpisodes(service *youtube.Service, videoIds []string, missingVideos []models.PodcastEpisode) []models.PodcastEpisode {
	videoCall := service.Videos.List([]string{"id,snippet,contentDetails"})
	videoCall = videoCall.Id(videoIds...)
	videoCall = videoCall.MaxResults(int64(len(videoIds)))
	videoResponse, err := videoCall.Do()
	if err != nil {
		log.Error(err)
		return missingVideos
	}

	for _, item := range videoResponse.Items {
		if item.Id != "" {
			duration, err := ParseDuration(item.ContentDetails.Duration)
			if err != nil {
				log.Error(err)
				continue
			}
			missingVideos = append(missingVideos, models.NewPodcastEpisodeFromSearch(item, duration))
		}
	}
	return missingVideos
}

func ParseDuration(durationStr string) (time.Duration, error) {
	// Remove the 'PT' prefix from the duration string
	durationStr = strings.Replace(durationStr, "PT", "", 1)

	// Replace 'H' with 'h', 'M' with 'm', and 'S' with 's'
	durationStr = strings.Replace(durationStr, "H", "h", 1)
	durationStr = strings.Replace(durationStr, "M", "m", 1)
	durationStr = strings.Replace(durationStr, "S", "s", 1)

	// Parse the duration string
	return time.ParseDuration(durationStr)
}

// Remove unavailable youtube videos used during the RSS feed generation
func cleanPlaylistItems(item *youtube.PlaylistItem) *youtube.PlaylistItem {
	unavailableStatuses := map[string]bool{
		"private":                  true,
		"unlisted":                 true,
		"privacyStatusUnspecified": true,
	}
	if item.Status != nil {
		if !unavailableStatuses[item.Status.PrivacyStatus] {
			return item
		}
	}

	return nil
}

func GetYoutubeVideo(youtubeVideoId string) (string, <-chan struct{}) {
	mutex, ok := youtubeVideoMutexes.Load(youtubeVideoId)
	if !ok {
		mutex = &sync.Mutex{}
		youtubeVideoMutexes.Store(youtubeVideoId, mutex)
	}

	mutex.(*sync.Mutex).Lock()

	// Check if the file is already being processed
	filePath := "/config/audio/" + youtubeVideoId + ".m4a"
	if _, err := os.Stat(filePath); err == nil {
		mutex.(*sync.Mutex).Unlock()
		return youtubeVideoId, make(chan struct{})
	}

	// If not, proceed with the download
	youtubeVideoId = strings.TrimSuffix(youtubeVideoId, ".m4a")
	ytdlp.Install(context.TODO(), nil)

	categories := os.Getenv("SPONSORBLOCK_CATEGORIES")
	if categories == "" {
		categories = "sponsor"
	}
	categories = strings.TrimSpace(categories)

	dl := ytdlp.New().
		NoProgress().
		FormatSort("ext::m4a").
		SponsorblockRemove(categories).
		ExtractAudio().
		NoPlaylist().
		FFmpegLocation("/usr/bin/ffmpeg").
		Continue().
		Paths("/config/audio").
		ProgressFunc(500*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			fmt.Printf(
				"%s @ %s [eta: %s] :: %s\n",
				prog.Status,
				prog.PercentString(),
				prog.ETA(),
				prog.Filename,
			)
		}).
		Output(youtubeVideoId + ".%(ext)s")

	cookiesFile := strings.TrimSpace(os.Getenv("COOKIES_FILE"))
	if cookiesFile != "" {
		dl.Cookies("/config/" + cookiesFile)
	}

	done := make(chan struct{})
	go func() {
		r, err := dl.Run(context.TODO(), youtubeVideoUrl+youtubeVideoId)
		if err != nil {
			log.Errorf("Error downloading YouTube video: %v", err)
		}
		if r.ExitCode != 0 {
			log.Errorf("YouTube video download failed with exit code %d", r.ExitCode)
		}
		mutex.(*sync.Mutex).Unlock()

		close(done)
	}()

	return youtubeVideoId, done
}

func setupYoutubeService() *youtube.Service {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatalf("GOOGLE_API_KEY is not set")
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Errorf("Error creating new YouTube client: %v", err)
	}
	if service == nil {
		log.Errorf("Failed to create YouTube service: %v", err)
	}
	return service
}

package services

import (
	"context"
	"fmt"
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
func getYoutubePlaylistData(youtubePlaylistId string) []*youtube.PlaylistItem {

	log.Info("[RSS FEED] Getting youtube data...")
	ctx := context.Background()

	service, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	continue_requesting_playlist_items := true
	allItems := []*youtube.PlaylistItem{}
	pageToken := "first_call"
	for continue_requesting_playlist_items {
		call := service.PlaylistItems.List([]string{"snippet", "status"}).
			PlaylistId(youtubePlaylistId).
			MaxResults(*maxResults)
		if pageToken != "first_call" {
			call.PageToken(pageToken)
		}

		response, ytAgainErr := call.Do()
		if response.HTTPStatusCode != http.StatusOK {
			log.Fatalf("YouTube API returned status code %v", response.HTTPStatusCode)
		}

		pageToken = response.NextPageToken
		if ytAgainErr != nil {
			log.Fatalf("Error calling YouTube API: %v", ytAgainErr)
		}
		allItems = append(allItems, response.Items...)
		if response.NextPageToken == "" {
			continue_requesting_playlist_items = false
		}
	}
	return allItems
}

func getChannelProfilePicture(channelID string) string {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		log.Error(err)
	}

	channelCall := youtubeService.Channels.List([]string{"brandingSettings", "snippet"})
	channelCall = channelCall.Id(channelID)

	channelResponse, err := channelCall.Do()
	if err != nil {
		log.Error(err)
	}

	if len(channelResponse.Items) == 0 {
		log.Error("channel not found")
	}

	channel := channelResponse.Items[0]
	profilePictureURL := channel.BrandingSettings.Image.BannerExternalUrl

	log.Info("[RSS FEED] Channel profile picture: ", profilePictureURL)
	return profilePictureURL
}

func getChannelMetadataAndVideos(channelID string) (ChannelMetadata, []VideoMetadata, error) {
	log.Info("[RSS FEED] Getting channel data...")
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
	if err != nil {
		return ChannelMetadata{}, nil, err
	}

	channelCall := youtubeService.Channels.List([]string{"snippet", "statistics", "contentDetails"})
	channelCall = channelCall.Id(channelID)

	channelResponse, err := channelCall.Do()
	if err != nil {
		return ChannelMetadata{}, nil, err
	}

	if len(channelResponse.Items) == 0 {
		log.Error("channel not found")
	}

	channelMetadata := ChannelMetadata{
		Title:        channelResponse.Items[0].Snippet.Title,
		Description:  channelResponse.Items[0].Snippet.Description,
		ThumbnailURL: channelResponse.Items[0].Snippet.Thumbnails.Default.Url,
		PublishedAt:  channelResponse.Items[0].Snippet.PublishedAt,
		Id:           channelResponse.Items[0].Id,
	}

	videoMetadata := make([]VideoMetadata, 0)

	nextPageToken := ""
	for {
		videoCall := youtubeService.Search.List([]string{"id,snippet"})
		videoCall = videoCall.ChannelId(channelID)
		videoCall = videoCall.Order("date")
		videoCall = videoCall.MaxResults(50)
		videoCall = videoCall.PageToken(nextPageToken)

		videoResponse, err := videoCall.Do()
		if err != nil {
			return ChannelMetadata{}, nil, err
		}

		for _, item := range videoResponse.Items {
			videoMetadata = append(videoMetadata, VideoMetadata{
				Title:        item.Snippet.Title,
				Description:  item.Snippet.Description,
				ThumbnailURL: item.Snippet.Thumbnails.Default.Url,
				VideoID:      item.Id.VideoId,
				PublishedAt:  item.Snippet.PublishedAt,
			})
		}

		if videoResponse.NextPageToken == "" {
			break
		}
		nextPageToken = videoResponse.NextPageToken
	}

	return channelMetadata, videoMetadata, nil
}

type ChannelMetadata struct {
	Title        string
	Description  string
	ThumbnailURL string
	PublishedAt  string
	Id           string
}

type VideoMetadata struct {
	Title        string
	Description  string
	ThumbnailURL string
	VideoID      string
	PublishedAt  string
}

// Remove unavailable youtube videos used during the RSS feed generation
func cleanPlaylistItems(items []*youtube.PlaylistItem) []*youtube.PlaylistItem {
	unavailableStatuses := map[string]bool{
		"private":                  true,
		"unlisted":                 true,
		"privacyStatusUnspecified": true,
	}
	cleanItems := make([]*youtube.PlaylistItem, 0)
	for _, item := range items {
		if item.Status != nil {
			if !unavailableStatuses[item.Status.PrivacyStatus] {
				cleanItems = append(cleanItems, item)
			}
		}
	}
	return cleanItems
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

	dl := ytdlp.New().
		NoProgress().
		FormatSort("ext::m4a").
		SponsorblockRemove("sponsor").
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

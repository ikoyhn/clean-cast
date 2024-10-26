package services

import (
	"context"
	"fmt"
	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

const youtubeVideoUrl = "https://www.youtube.com/watch?v="

// Get all youtube playlist items and meta data for the RSS feed
func getYoutubeData(youtubePlaylistId string) []*youtube.PlaylistItem {

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

// Remove unavailable youtube videos used during the RSS feed generation
func cleanPlaylistItems(items []*youtube.PlaylistItem) []*youtube.PlaylistItem {
	var cleanItems []*youtube.PlaylistItem
	for _, item := range items {
		if !slices.Contains(UNAVAILABLE_STATUSES, item.Status.PrivacyStatus) {
			cleanItems = append(cleanItems, item)
		}
	}
	return cleanItems
}

// When client requests a podcast from the RSS feed use yt-dlp to download the file then serve it
func GetYoutubeVideo(youtubeVideoId string) (string, <-chan struct{}) {
	log.Info("[YTDLP] Checking for video...")
	filePath := "/config/" + youtubeVideoId + ".m4a"
	if _, err := os.Stat(filePath); err == nil {
		return youtubeVideoId, make(chan struct{})
	}
	log.Info("[YTDLP] Video not found, downloading...")
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
		Paths("/config").
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
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
			// handle error
		}
		if r.ExitCode != 0 {
			// handle error
		}
		close(done)
	}()

	return youtubeVideoId, done
}

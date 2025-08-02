package downloader

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/labstack/gommon/log"
	ytdlp "github.com/lrstanley/go-ytdlp"
)

var youtubeVideoMutexes = &sync.Map{}

const youtubeVideoUrl = "https://www.youtube.com/watch?v="

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
		FormatSort("ext::m4a[format_note*=original]").
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

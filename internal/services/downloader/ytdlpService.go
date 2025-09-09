package downloader

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/services/ntfy"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
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
	filePath := config.Config.AudioDir + youtubeVideoId + ".m4a"
	if _, err := os.Stat(filePath); err == nil {
		mutex.(*sync.Mutex).Unlock()
		return youtubeVideoId, make(chan struct{})
	}

	// If not, proceed with the download
	youtubeVideoId = strings.TrimSuffix(youtubeVideoId, ".m4a")
	ytdlp.Install(context.TODO(), nil)

	categories := config.Config.SponsorBlockCategories
	if categories == "" {
		categories = "sponsor"
	}
	categories = strings.TrimSpace(categories)

	var etaNotified uint32 = 0
	dl := ytdlp.New().
		NoProgress().
		FormatSort("ext::m4a[format_note*=original]").
		SponsorblockRemove(categories).
		ExtractAudio().
		NoPlaylist().
		FFmpegLocation("/usr/bin/ffmpeg").
		Continue().
		Paths(config.Config.AudioDir).
		ProgressFunc(4000*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			ytdlpProgress(&etaNotified, prog, youtubeVideoId)
		}).
		Output(youtubeVideoId + ".%(ext)s")

	if config.Config.CookiesFile != "" {
		dl.Cookies(config.Config.CookiesFile)
	}
	if config.Config.YtdlpExtractorArgs != "" {
		dl.ExtractorArgs(config.Config.YtdlpExtractorArgs)
	}

	done := make(chan struct{})
	go func() {
		r, dlErr := dl.Run(context.TODO(), youtubeVideoUrl+youtubeVideoId)

		if r.ExitCode != 0 {
			file, err := os.Open(path.Join(config.Config.AudioDir, youtubeVideoId+".m4a"))
			if file != nil || err == nil {
				ntfy.SendNotification("Download completed!", "Clean Cast - Success")
				log.Warn("Download exited with non-zero code, but file exists: ", filePath)
			} else {
				if dlErr != nil {
					ntfy.SendNotification("Download failed!", "Clean Cast - Error")
					log.Errorf("Error downloading YouTube video: %v", dlErr)
				}
			}
		} else {
			log.Infof("%s Download completed successfully.", youtubeVideoId)
			ntfy.SendNotification("Download completed!", "Clean Cast - Success")

		}
		mutex.(*sync.Mutex).Unlock()

		close(done)
	}()

	return youtubeVideoId, done
}

func ytdlpProgress(etaNotified *uint32, prog ytdlp.ProgressUpdate, youtubeVideoId string) {
	fmt.Printf(
		"%s @ %s [eta: %s] :: %s\n",
		prog.Status,
		prog.PercentString(),
		prog.ETA(),
		prog.Filename,
	)

	if atomic.LoadUint32(etaNotified) == 0 {
		eta := prog.ETA()
		if eta > time.Duration(0) {
			if atomic.CompareAndSwapUint32(etaNotified, 0, 1) {
				msg := fmt.Sprintf("Downloading %s — %s @ %s (eta: %s)", youtubeVideoId, prog.Status, prog.PercentString(), eta)
				ntfy.SendNotification(msg, "Clean Cast")
			}
		}
	}
}

package downloader

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
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

func GetYoutubeVideo(youtubeVideoIdFile string) (string, <-chan struct{}) {
	mutex, ok := youtubeVideoMutexes.Load(youtubeVideoIdFile)
	if !ok {
		mutex = &sync.Mutex{}
		youtubeVideoMutexes.Store(youtubeVideoIdFile, mutex)
	}

	mutex.(*sync.Mutex).Lock()

	filePath := config.AppConfig.Setup.AudioDir + youtubeVideoIdFile
	if _, err := os.Stat(filePath); err == nil {
		mutex.(*sync.Mutex).Unlock()
		return youtubeVideoIdFile, make(chan struct{})
	}

	youtubeVideoId := strings.TrimSuffix(youtubeVideoIdFile, ".m4a")
	title := youtubeVideoId
	episode, err := database.GetEpisodeByVideoId(youtubeVideoId)
	if err != nil {
		log.Warnf("Error fetching YouTube video title: %v", err)
	}
	if episode != nil && episode.EpisodeName != "" {
		title = episode.EpisodeName
	}

	categories := config.AppConfig.Ytdlp.SponsorBlockCategories
	categories = strings.TrimSpace(categories)

	var etaNotified uint32 = 0
	dl := ytdlp.New().
		NoProgress().
		// prefer m4a audio with original format note, but accept other extensions as a fallback
		FormatSort("ext::m4a[format_note*=original]").
		SponsorblockRemove(categories).
		ExtractAudio().
		NoPlaylist().
		FFmpegLocation("/usr/bin/ffmpeg").
		Continue().
		Paths(config.AppConfig.Setup.AudioDir).
		ProgressFunc(4000*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			ytdlpProgress(&etaNotified, prog, title)
		}).
		Output(youtubeVideoId + ".%(ext)s")

	if config.AppConfig.Ytdlp.CookiesFile != "" {
		dl.Cookies(config.AppConfig.Ytdlp.CookiesFile)
	}
	if config.AppConfig.Ytdlp.YtdlpExtractorArgs != "" {
		dl.ExtractorArgs(config.AppConfig.Ytdlp.YtdlpExtractorArgs)
	}

	done := make(chan struct{})
	go func() {
		r, dlErr := dl.Run(context.TODO(), youtubeVideoUrl+youtubeVideoId)

		if r.ExitCode != 0 {
			file, err := os.Open(path.Join(config.AppConfig.Setup.AudioDir, youtubeVideoId+".m4a"))
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
			log.Infof("%s download completed successfully.", title)
			ntfy.SendNotification(fmt.Sprintf("%s download success!", title), "Clean Cast - Success")

		}
		mutex.(*sync.Mutex).Unlock()

		close(done)
	}()

	return youtubeVideoId, done
}

func ytdlpProgress(etaNotified *uint32, prog ytdlp.ProgressUpdate, title string) {
	fmt.Printf(
		"%s @ %s [eta: %s]\n",
		prog.Status,
		prog.PercentString(),
		prog.ETA(),
	)

	if atomic.LoadUint32(etaNotified) == 0 {
		eta := prog.ETA()
		if eta > time.Duration(0) {
			if atomic.CompareAndSwapUint32(etaNotified, 0, 1) {
				msg := fmt.Sprintf("%s — %s @ %s (eta: %s)", title, prog.Status, prog.PercentString(), eta)
				ntfy.SendNotification(msg, "Clean Cast")
			}
		}
	}
}

package downloader

import (
	"context"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/services/ntfy"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/labstack/gommon/log"
	"github.com/lrstanley/go-ytdlp"
)

var youtubeVideoMutexes = &sync.Map{}

const youtubeVideoUrl = "https://www.youtube.com/watch?v="

func GetYoutubeVideo(youtubeVideoId string) <-chan struct{} {
	mutex, ok := youtubeVideoMutexes.Load(youtubeVideoId)
	if !ok {
		mutex = &sync.Mutex{}
		youtubeVideoMutexes.Store(youtubeVideoId, mutex)
	}

	mutex.(*sync.Mutex).Lock()

	if database.FileExistsWithId(config.AppConfig.Setup.AudioDir, youtubeVideoId) {
		mutex.(*sync.Mutex).Unlock()
		return make(chan struct{})
	}

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
		Format("bestaudio[ext=m4a]/bestaudio/best").
		SponsorblockRemove(categories).
		RemoteComponents("COMPONENTS").
		JsRuntimes("bun").
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
			if database.FileExistsWithId(config.AppConfig.Setup.AudioDir, youtubeVideoId) {
				ntfy.SendNotification("Download completed!", "Clean Cast - Success")
				log.Warn("Download exited with non-zero code, but file exists: ", youtubeVideoId)
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

	return done
}

func ytdlpProgress(etaNotified *uint32, prog ytdlp.ProgressUpdate, title string) {
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
				msg := fmt.Sprintf("%s — %s @ %s (eta: %s)", title, prog.Status, prog.PercentString(), eta)
				ntfy.SendNotification(msg, "Clean Cast")
			}
		}
	}
}

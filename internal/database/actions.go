package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"
	"time"

	log "github.com/labstack/gommon/log"
)

func UpdateEpisodePlaybackHistory(youtubeVideoId string) {
	db.Model(&models.EpisodePlaybackHistory{}).Where("youtube_video_id = ?", youtubeVideoId).Update("last_access_date", time.Now())
}

func DeletePodcastCronJob() {
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour)
	db.Find(&models.EpisodePlaybackHistory{}, "last_access_date < ?", oneWeekAgo).Delete(&models.EpisodePlaybackHistory{})
}

func TrackEpisodeFiles() {
	files, err := os.ReadDir("/config/audio/")
	if err != nil {
		log.Fatal(err)
	}

	dbFiles := make([]string, 0)
	db.Model(&models.EpisodePlaybackHistory{}).Pluck("YoutubeVideoId", &dbFiles)

	missingFiles := make([]string, 0)
	for _, file := range files {
		filename := file.Name()
		found := false
		for _, dbFile := range dbFiles {
			if dbFile == filename[:len(filename)-4] {
				found = true
				break
			}
		}
		if !found {
			missingFiles = append(missingFiles, filename)
		}
	}

	for _, filename := range missingFiles {
		id := filename[:len(filename)-4]
		db.Create(&models.EpisodePlaybackHistory{YoutubeVideoId: id, LastAccessDate: time.Now()})
	}
}

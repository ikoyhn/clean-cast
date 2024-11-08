package database

import (
	"ikoyhn/podcast-sponsorblock/internal/models"
	"os"
	"time"

	log "github.com/labstack/gommon/log"
)

func UpdateEpisodePlaybackHistory(youtubeVideoId string) {
	log.Info("[DB] Updating episode playback history...")
	db.Model(&models.EpisodePlaybackHistory{}).Where("youtube_video_id = ?", youtubeVideoId).Update("last_access_date", time.Now().Unix())
}

func DeletePodcastCronJob() {
	oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()

	log.Info("[DB] Deleting old episode files...")
	var histories []models.EpisodePlaybackHistory
	db.Where("last_access_date < ?", oneWeekAgo).Find(&histories)

	for _, history := range histories {
		os.Remove("/config/audio/" + history.YoutubeVideoId + ".m4a")
		db.Delete(&history)
		log.Info("[DB] Deleted old episode playback history... " + history.YoutubeVideoId)
	}
}

func TrackEpisodeFiles() {
	log.Info("[DB] Tracking existing episode files...")
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
		db.Create(&models.EpisodePlaybackHistory{YoutubeVideoId: id, LastAccessDate: time.Now().Unix()})
	}
}

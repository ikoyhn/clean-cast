package sponsorblock

import (
	"encoding/json"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/constants"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/logger"
)

const SPONSORBLOCK_API_URL = "https://sponsor.ajay.app/api/skipSegments?videoID="

// Package-level HTTP client with connection pooling and timeout
var httpClient = &http.Client{
	Timeout: constants.RequestTimeout * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

func DeterminePodcastDownload(youtubeVideoId string) (bool, float64) {
	episodeHistory := database.GetEpisodePlaybackHistory(youtubeVideoId)

	updatedSkippedTime := TotalSponsorTimeSkipped(youtubeVideoId)
	if episodeHistory == nil {
		return true, updatedSkippedTime
	}

	if math.Abs(episodeHistory.TotalTimeSkipped-updatedSkippedTime) > constants.SponsorBlockThreshold {
		os.Remove(config.Config.AudioDir + youtubeVideoId + ".m4a")
		logger.Logger.Debug().Msg("[SponsorBlock] Updating downloaded episode with new sponsor skips...")
		return true, updatedSkippedTime
	}

	return false, updatedSkippedTime
}

func TotalSponsorTimeSkipped(youtubeVideoId string) float64 {
	logger.Logger.Debug().Msg("[SponsorBlock] Looking up podcast in SponsorBlock API...")
	endURL := SPONSORBLOCK_API_URL + youtubeVideoId

	if categories := getCategories(); categories != nil {
		for _, category := range categories {
			endURL += "&category=" + strings.TrimSpace(category)
		}
	}

	resp, err := httpClient.Get(endURL)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		logger.Logger.Warn().Msgf("Video not found on SponsorBlock API: %s", youtubeVideoId)
		return 0
	}

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		logger.Logger.Error(bodyErr)
		return 0
	}
	sponsorBlockResponse, marshErr := unmarshalSponsorBlockResponse(body)
	if marshErr != nil {
		logger.Logger.Error(marshErr)
		return 0
	}

	totalTimeSkipped := calculateSkippedTime(sponsorBlockResponse)

	return totalTimeSkipped
}

func unmarshalSponsorBlockResponse(data []byte) ([]SponsorBlockResponse, error) {
	var res []SponsorBlockResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return []SponsorBlockResponse{}, err
	}

	return res, nil
}

func calculateSkippedTime(segments []SponsorBlockResponse) float64 {
	skippedTime := float64(0)
	prevStopTime := float64(0)

	for _, segment := range segments {
		startTime := segment.Segment[0]
		stopTime := segment.Segment[1]

		if startTime > prevStopTime {
			skippedTime += stopTime - startTime
		} else if stopTime > prevStopTime {
			skippedTime += stopTime - prevStopTime
		}

		if stopTime > prevStopTime {
			prevStopTime = stopTime
		}
	}

	return skippedTime
}

func getCategories() []string {
	if config.Config.SponsorBlockCategories == "" {
		return nil
	}
	return strings.Split(config.Config.SponsorBlockCategories, ",")
}

type SponsorBlockResponse struct {
	Segment       []float64 `json:"segment"`
	UUID          string    `json:"UUID"`
	Category      string    `json:"category"`
	VideoDuration float64   `json:"videoDuration"`
	ActionType    string    `json:"actionType"`
	Locked        int16     `json:"locked"`
	Votes         int16     `json:"votes"`
	Description   string    `json:"description"`
}

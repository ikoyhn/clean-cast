package services

import (
	"encoding/json"
	log "github.com/labstack/gommon/log"
	"io"
	"net/http"
)

const SPONSORBLOCK_API_URL = "https://sponsor.ajay.app/api/skipSegments?videoID="

func TotalSponsorTimeSkipped(youtubeVideoId string) float64 {
	log.Info("[SponsorBlock] Looking up podcast in SponsorBlock API...")
	resp, err := http.Get(SPONSORBLOCK_API_URL + youtubeVideoId)
	if err != nil {
		log.Fatal(err)
	}
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		log.Fatal(bodyErr)
	}
	sponsorBlockResponse, marshErr := unmarshalSponsorBlockResponse(body)
	if marshErr != nil {
		log.Fatal(marshErr)
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
		} else {
			skippedTime += stopTime - prevStopTime
		}

		prevStopTime = stopTime
	}

	return skippedTime
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

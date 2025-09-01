package apple

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/labstack/gommon/log"
)

const ITUNES_SEARCH_URL = "https://itunes.apple.com/search?term=%s&limit=1&media=podcast&callback="

func GetApplePodcastData(podcastName string) LookupResponse {
	log.Debug("[RSS FEED] Looking up podcast in Apple Search API...")
	resp, err := http.Get(fmt.Sprintf(ITUNES_SEARCH_URL, strings.ReplaceAll(podcastName, " ", "")))
	if err != nil {
		log.Error(err)
	}
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		log.Error(bodyErr)
	}
	lookupResponse, marshErr := unmarshalAppleLookupResponse(body)
	if marshErr != nil {
		log.Error(marshErr)
	}
	return lookupResponse
}

func unmarshalAppleLookupResponse(data []byte) (LookupResponse, error) {
	var res LookupResponse

	if err := json.Unmarshal(data, &res); err != nil {
		return LookupResponse{}, err
	}

	return res, nil
}

// If the Apple API search call returns more than one result find the one with the closest number of episodes
func findClosestResult(results []AppleResult, target int) AppleResult {
	var closest AppleResult
	minDiff := math.MaxInt32

	for _, result := range results {
		diff := int(math.Abs(float64(result.TrackCount - target)))
		if diff < minDiff {
			minDiff = diff
			closest = result
		}
	}
	return closest
}

func getAppleData(channelTitle string, numOfVideos int) AppleResult {
	itunesResponse := GetApplePodcastData(channelTitle)
	closestApplePodcastData := findClosestResult(itunesResponse.Results, numOfVideos)
	return closestApplePodcastData
}

type LookupResponse struct {
	// ResultCount contains info about total found results number.
	ResultCount int64 `json:"resultCount"`
	// Results is an array with all found results.
	Results []AppleResult `json:"results"`
}

type AppleResult struct {
	CollectionId          int    `json:"collectionId"`
	TrackCount            int    `json:"trackCount"`
	PrimaryGenreName      string `json:"primaryGenreName"`
	ContentAdvisoryRating string `json:"contentAdvisoryRating"`
	ArtworkUrl100         string `json:"artworkUrl100"`
	ReleaseDate           string `json:"releaseDate"`
	TrackName             string `json:"trackName"`
	ArtistName            string `json:"artistName"`
}

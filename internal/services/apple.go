package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	log "github.com/labstack/gommon/log"
)

const itunesSearchUrl = "https://itunes.apple.com/search?term=%s&limit=1&media=podcast&callback="

// Apple API lookup for podcast metadata
func GetApplePodcastData(podcastName string) LookupResponse {
	log.Info("[RSS FEED] Looking up podcast in Apple Search API...")
	resp, err := http.Get(fmt.Sprintf(itunesSearchUrl, strings.ReplaceAll(podcastName, " ", "")))
	if err != nil {
		log.Fatal(err)
	}
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		log.Fatal(bodyErr)
	}
	lookupResponse, marshErr := unmarshalLookupResponse(body)
	if marshErr != nil {
		log.Fatal(marshErr)
	}
	return lookupResponse
}

func unmarshalLookupResponse(data []byte) (LookupResponse, error) {
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

package apple

import (
	"encoding/json"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/store"
	"io"
	"net/http"
	"strings"

	log "github.com/labstack/gommon/log"
)

const ITUNES_TOP_PODCASTS_URL = "https://itunes.apple.com/gb/rss/toppodcasts/limit=%d/json"
const ITUNES_SEARCH_URL = "https://itunes.apple.com/search?term=%s&limit=1&media=podcast&callback="

type TopPodcastResponse struct {
	Feed struct {
		Entries []struct {
			ID struct {
				Label string `json:"label"`
			}
			Name struct {
				Label string `json:"label"`
			} `json:"im:name"`
			Images []struct {
				Label string `json:"label"`
			} `json:"im:image"`
			Category struct {
				Attributes struct {
					Label string `json:"label"`
				} `json:"attributes"`
			} `json:"category"`
			Summary struct {
				Label string `json:"label"`
			} `json:"summary"`
		} `json:"entry"`
	} `json:"feed"`
}

func GetTopPodcasts(limit int) ([]models.TopPodcast, error) {
	log.Debug("[DASHBOARD] Getting Top Podcasts...")
	val, ok := store.GetValue(fmt.Sprintf(ITUNES_TOP_PODCASTS_URL, limit))
	if ok {
		return val.([]models.TopPodcast), nil
	}

	resp, err := http.Get(fmt.Sprintf(ITUNES_TOP_PODCASTS_URL, limit))
	if err != nil {
		return []models.TopPodcast{}, err
	}
	defer resp.Body.Close()

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return []models.TopPodcast{}, bodyErr
	}
	var lookupResponse TopPodcastResponse
	marshErr := json.Unmarshal(body, &lookupResponse)
	if marshErr != nil {
		if syntaxErr, ok := marshErr.(*json.SyntaxError); ok {
			log.Errorf("JSON syntax error at offset %d: %s", syntaxErr.Offset, syntaxErr)
		} else {
			log.Errorf("JSON unmarshal error: %s", marshErr)
		}
		return []models.TopPodcast{}, marshErr
	}

	topPodcasts := make([]models.TopPodcast, len(lookupResponse.Feed.Entries))
	for i, result := range lookupResponse.Feed.Entries {
		topPodcasts[i] = models.TopPodcast{
			Id:          result.ID.Label,
			Image:       result.Images[len(result.Images)-1].Label,
			Title:       strings.ReplaceAll(result.Name.Label, "'", ""),
			Category:    result.Category.Attributes.Label,
			Description: result.Summary.Label,
		}
	}

	store.StoreValue(fmt.Sprintf(ITUNES_TOP_PODCASTS_URL, limit), topPodcasts)
	return topPodcasts, nil
}

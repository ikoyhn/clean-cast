package services

import (
	"encoding/xml"
	"fmt"
	podcastRss "github.com/eduncan911/podcast"
	"github.com/labstack/echo/v4"
	log "github.com/labstack/gommon/log"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const rssUrl = "/rss/youtubePlaylistId="

func GenerateRssFeed(podcast models.Podcast, c echo.Context) error {
	log.Info("[RSS FEED] Generating RSS Feed with Youtube and Apple metadata")

	atomLink := &podcastRss.AtomLink{
		XMLName: xml.Name{Space: "http://www.w3.org/2005/Atom", Local: "link"},
		HREF:    "https://example.com",
		Rel:     "self",
		Type:    "application/rss+xml",
	}

	now := time.Now()
	ytPodcast := podcastRss.New(podcast.PodcastName, "https://www.youtube.com/playlist?list="+podcast.YoutubePodcastId, podcast.Description, &now, &now)
	ytPodcast.AddImage(transformArtworkURL(podcast.ImageUrl, 3000, 3000))
	ytPodcast.AtomLink = atomLink
	ytPodcast.AddCategory(podcast.Category, []string{""})

	if podcast.PodcastEpisodes != nil {
		for _, podcastEpisode := range podcast.PodcastEpisodes {
			enclosure := podcastRss.Enclosure{
				URL:    handler(c.Request()) + "/media/" + podcastEpisode.YoutubeVideoId + ".m4a",
				Length: 0,
				Type:   podcastRss.M4A,
			}

			var builder strings.Builder
			xml.EscapeText(&builder, []byte(podcastEpisode.EpisodeDescription))
			escapedDescription := builder.String()

			podcastItem := podcastRss.Item{
				Title:       podcastEpisode.EpisodeName,
				Description: escapedDescription,
				GUID:        podcastEpisode.YoutubeVideoId,
				Category:    podcast.Category,
				Enclosure:   &enclosure,
				PubDate:     &now,
			}
			ytPodcast.AddItem(podcastItem)
		}

		err := ytPodcast.Encode(c.Response().Writer)
		if err != nil {
			return err
		}
	}
	return c.NoContent(http.StatusOK)
}

func handler(r *http.Request) string {
	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	host := r.Host
	url := scheme + "://" + host
	return url
}

func transformArtworkURL(artworkURL string, newHeight int, newWidth int) string {
	parsedURL, err := url.Parse(artworkURL)
	if err != nil {
		return ""
	}

	pathComponents := strings.Split(parsedURL.Path, "/")
	lastComponent := pathComponents[len(pathComponents)-1]
	newFilename := fmt.Sprintf("%dx%d%s", newHeight, newWidth, filepath.Ext(lastComponent))
	pathComponents[len(pathComponents)-1] = newFilename
	newPath := strings.Join(pathComponents, "/")

	newURL := parsedURL
	newURL.Path = newPath

	return newURL.String()
}

package services

import (
	"encoding/xml"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/labstack/gommon/log"
)

func GenerateRssFeed(podcast models.Podcast, host string) []byte {
	log.Info("[RSS FEED] Generating RSS Feed with Youtube and Apple metadata")

	now := time.Now()
	ytPodcast := New(podcast.PodcastName, "https://www.youtube.com/playlist?list="+podcast.Id, podcast.Description, &now)
	ytPodcast.AddImage(transformArtworkURL(podcast.ImageUrl, 1000, 1000))
	ytPodcast.AddCategory(podcast.Category, []string{""})
	ytPodcast.Docs = "http://www.rssboard.org/rss-specification"
	ytPodcast.IAuthor = podcast.ArtistName

	if podcast.PodcastEpisodes != nil {
		for _, podcastEpisode := range podcast.PodcastEpisodes {
			if podcastEpisode.EpisodeName == "Private video" || podcastEpisode.EpisodeDescription == "This video is private." {
				continue
			}
			mediaUrl := host + "/media/" + podcastEpisode.YoutubeVideoId + ".m4a"

			if os.Getenv("TOKEN") != "" {
				mediaUrl = mediaUrl + "?token=" + os.Getenv("TOKEN")
			}
			enclosure := Enclosure{
				URL:    mediaUrl,
				Length: 0,
				Type:   M4A,
			}

			var builder strings.Builder
			xml.EscapeText(&builder, []byte(podcastEpisode.EpisodeDescription))
			escapedDescription := builder.String()

			parseTime := parseTimeFromString(podcastEpisode.PublishedDate)

			podcastItem := Item{
				Title:       podcastEpisode.EpisodeName,
				Description: escapedDescription,
				GUID: struct {
					Value       string `xml:",chardata"`
					IsPermaLink bool   `xml:"isPermaLink,attr"`
				}{
					Value:       podcastEpisode.YoutubeVideoId,
					IsPermaLink: false,
				},
				Enclosure: &enclosure,
				PubDate:   &parseTime,
			}
			ytPodcast.AddItem(podcastItem)
		}
	}

	return ytPodcast.Bytes()
}

func parseTimeFromString(date string) time.Time {
	parseTime, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Error("Failed to parse time: " + date)
	}
	return parseTime
}

func transformArtworkURL(artworkURL string, newHeight int, newWidth int) string {
	parsedURL, err := url.Parse(artworkURL)
	if err != nil {
		return ""
	}

	log.Debug("[RSS FEED] Transforming image url...", artworkURL)
	pathComponents := strings.Split(parsedURL.Path, "/")
	lastComponent := pathComponents[len(pathComponents)-1]
	ext := filepath.Ext(lastComponent)
	if ext == "" {
		log.Debug("[RSS FEED] No file extension found, returning original URL")
		return artworkURL
	}

	newFilename := fmt.Sprintf("%dx%d%s", newHeight, newWidth, ext)
	pathComponents[len(pathComponents)-1] = newFilename
	newPath := strings.Join(pathComponents, "/")

	newURL := url.URL{
		Scheme:   parsedURL.Scheme,
		Host:     parsedURL.Host,
		Path:     newPath,
		RawQuery: parsedURL.RawQuery,
		Fragment: parsedURL.Fragment,
	}

	log.Debug("[RSS FEED] New image url: ", newURL.String())

	return newURL.String()
}

package services

import (
	"encoding/xml"
	"fmt"
	log "github.com/labstack/gommon/log"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GenerateRssFeed(podcast models.Podcast, appleData AppleResult, host string) []byte {
	log.Info("[RSS FEED] Generating RSS Feed with Youtube and Apple metadata")

	now := time.Now()
	ytPodcast := New(podcast.PodcastName, "https://www.youtube.com/playlist?list="+podcast.YoutubePodcastId, podcast.Description, &now)
	ytPodcast.AddImage(transformArtworkURL(podcast.ImageUrl, 1000, 1000))
	ytPodcast.AddCategory(podcast.Category, []string{""})
	ytPodcast.Docs = "http://www.rssboard.org/rss-specification"
	ytPodcast.IAuthor = appleData.ArtistName

	if podcast.PodcastEpisodes != nil {
		for _, podcastEpisode := range podcast.PodcastEpisodes {
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
		log.Fatal("Failed to parse time: " + date)
	}
	return parseTime
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

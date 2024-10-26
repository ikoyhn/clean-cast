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


func GenerateRssFeed(podcast models.Podcast, host string) []byte {
	log.Info("[RSS FEED] Generating RSS Feed with Youtube and Apple metadata")

	now := time.Now()
	ytPodcast := New(podcast.PodcastName, "https://www.youtube.com/playlist?list="+podcast.YoutubePodcastId, podcast.Description, &now, &now)
	ytPodcast.AddImage(transformArtworkURL(podcast.ImageUrl, 3000, 3000))
	ytPodcast.AddCategory(podcast.Category, []string{""})
	ytPodcast.IExplicit = "true"
	ytPodcast.Docs = "http://www.rssboard.org/rss-specification"

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

			podcastItem := Item{
				Title:       podcastEpisode.EpisodeName,
				Description: escapedDescription,
				GUID:        podcastEpisode.YoutubeVideoId,
				Category:    podcast.Category,
				Enclosure:   &enclosure,
				PubDate:     &now,
			}
			ytPodcast.AddItem(podcastItem)
		}
	}

	return ytPodcast.Bytes()
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

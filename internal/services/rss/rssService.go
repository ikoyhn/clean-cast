package rss

import (
	"encoding/xml"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/generator"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	log "github.com/labstack/gommon/log"
)

func GenerateRssFeed(podcast models.Podcast, host string, podcastType enum.PodcastType) []byte {
	log.Info("[RSS FEED] Generating RSS Feed...")

	podcastLink := "https://www.youtube.com/playlist?list=" + podcast.Id

	if podcastType == enum.CHANNEL {
		podcastLink = "https://www.youtube.com/channel/" + podcast.Id
	}

	now := time.Now()
	ytPodcast := generator.New(podcast.PodcastName, podcastLink, podcast.Description, &now)
	ytPodcast.AddImage(transformArtworkURL(podcast.ImageUrl, 1000, 1000))
	ytPodcast.AddCategory(podcast.Category, []string{""})
	ytPodcast.Docs = "http://www.rssboard.org/rss-specification"
	ytPodcast.IAuthor = podcast.ArtistName

	if podcast.PodcastEpisodes != nil {
		for _, podcastEpisode := range podcast.PodcastEpisodes {
			if (podcastEpisode.Type == "CHANNEL" && podcastEpisode.Duration.Seconds() < 120) || podcastEpisode.EpisodeName == "Private video" || podcastEpisode.EpisodeDescription == "This video is private." {
				continue
			}
			mediaUrl := host + "/media/" + podcastEpisode.YoutubeVideoId + ".m4a"

			if config.AppConfig.Authentication.Token != "" {
				mediaUrl = mediaUrl + "?token=" + config.AppConfig.Authentication.Token
			}
			enclosure := generator.Enclosure{
				URL:    mediaUrl,
				Length: 0,
				Type:   generator.M4A,
			}

			var builder strings.Builder
			xml.EscapeText(&builder, []byte(podcastEpisode.EpisodeDescription))
			escapedDescription := builder.String()

			podcastItem := generator.Item{
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
				PubDate:   &podcastEpisode.PublishedDate,
			}

			// Add image if available
			if podcastEpisode.ImageUrl != "" {
				podcastItem.IImage = struct {
					Href string `xml:"href,attr"`
				}{
					Href: podcastEpisode.ImageUrl,
				}
			}

			ytPodcast.AddItem(podcastItem)
		}
	}

	return ytPodcast.Bytes()
}

func BuildPodcast(podcast models.Podcast, allItems []models.PodcastEpisode) models.Podcast {
	podcast.PodcastEpisodes = allItems
	return podcast
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

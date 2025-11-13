package rss

import (
	"encoding/xml"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/cache"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/enum"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/channel"
	"ikoyhn/podcast-sponsorblock/internal/services/downloader"
	"ikoyhn/podcast-sponsorblock/internal/services/generator"
	"ikoyhn/podcast-sponsorblock/internal/services/youtube"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/logger"
)

func GenerateRssFeed(podcast models.Podcast, host string, podcastType enum.PodcastType, audioFormat downloader.AudioFormat) []byte {
	logger.Logger.Info().Msg("[RSS FEED] Generating RSS Feed...")

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

	// Convert audioFormat to generator.EnclosureType
	enclosureType := getEnclosureType(audioFormat)

	if podcast.PodcastEpisodes != nil {
		for _, podcastEpisode := range podcast.PodcastEpisodes {
			if (podcastEpisode.Type == "CHANNEL" && podcastEpisode.Duration.Seconds() < 120) || podcastEpisode.EpisodeName == "Private video" || podcastEpisode.EpisodeDescription == "This video is private." {
				continue
			}
			mediaUrl := host + "/media/" + podcastEpisode.YoutubeVideoId + audioFormat.Extension

			// Add format parameter to URL if not using default format
			if audioFormat.Format != config.Config.AudioFormat {
				mediaUrl = mediaUrl + "?format=" + audioFormat.Format
				if config.Config.Token != "" {
					mediaUrl = mediaUrl + "&token=" + config.Config.Token
				}
			} else if config.Config.Token != "" {
				mediaUrl = mediaUrl + "?token=" + config.Config.Token
			}

			enclosure := generator.Enclosure{
				URL:    mediaUrl,
				Length: 0,
				Type:   enclosureType,
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
			ytPodcast.AddItem(podcastItem)
		}
	}

	return ytPodcast.Bytes()
}

// getEnclosureType converts AudioFormat to generator.EnclosureType
func getEnclosureType(audioFormat downloader.AudioFormat) generator.EnclosureType {
	switch audioFormat.Format {
	case "mp3":
		return generator.MP3
	case "m4a":
		return generator.M4A
	case "opus":
		// For opus, we'll use M4A as fallback since generator doesn't define OPUS
		// The MIME type will be correct via audioFormat.MimeType
		return generator.M4A
	default:
		return generator.M4A
	}
}

func BuildChannelRssFeed(channelId string, params *models.RssRequestParams, host string, audioFormat downloader.AudioFormat) []byte {
	logger.Logger.Info().Msg("[RSS FEED] Building rss feed for channel...")

	// Check cache first
	rssFeedCache := cache.GetRSSFeedCache()
	cacheParams := map[string]interface{}{"host": host, "params": params, "format": audioFormat.Format}
	if cachedFeed, found := rssFeedCache.Get("channel", channelId, cacheParams); found {
		logger.Logger.Info().Msg("[RSS FEED] Returning cached channel feed")
		return cachedFeed
	}

	service := youtube.SetupYoutubeService()

	podcast := youtube.GetChannelData(channelId, service, false)

	channel.GetChannelMetadataAndVideos(podcast.Id, service, params)
	episodes, err := database.GetPodcastEpisodesByPodcastId(podcast.Id, enum.CHANNEL)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("")
		return nil
	}

	// Apply filters and preferences
	episodes, err = database.ApplyFiltersAndPreferences(episodes, podcast.Id)
	if err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to apply filters and preferences")
		// Continue without filters rather than failing completely
	}

	podcastRss := BuildPodcast(podcast, episodes)
	rssFeed := GenerateRssFeed(podcastRss, host, enum.CHANNEL, audioFormat)

	// Store in cache
	rssFeedCache.Set("channel", channelId, cacheParams, rssFeed)

	return rssFeed
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

	logger.Logger.Debug().Msg("[RSS FEED] Transforming image url...", artworkURL)
	pathComponents := strings.Split(parsedURL.Path, "/")
	lastComponent := pathComponents[len(pathComponents)-1]
	ext := filepath.Ext(lastComponent)
	if ext == "" {
		logger.Logger.Debug().Msg("[RSS FEED] No file extension found, returning original URL")
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

	logger.Logger.Debug().Msg("[RSS FEED] New image url: ", newURL.String())

	return newURL.String()
}

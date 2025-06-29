package models

import (
	"time"

	log "github.com/labstack/gommon/log"
	"google.golang.org/api/youtube/v3"
)

type PodcastEpisode struct {
	Id                 int32         `gorm:"autoIncrement;primary_key;not null"`
	YoutubeVideoId     string        `json:"youtube_video_id" gorm:"index:youtubevideoid_type"`
	EpisodeName        string        `json:"episode_name"`
	EpisodeDescription string        `json:"episode_description"`
	PublishedDate      time.Time     `json:"published_date"`
	Type               string        `json:"type" gorm:"index:youtubevideoid_type_channelid_type"`
	PodcastId          string        `json:"podcast_id" gorm:"foreignkey:PodcastId;association_foreignkey:Id"`
	Duration           time.Duration `json:"duration"`
}

type Podcast struct {
	Id              string           `json:"id" gorm:"primary_key"`
	AppleId         string           `json:"apple_id"`
	PodcastName     string           `json:"podcast_name"`
	Description     string           `json:"description"`
	Category        string           `json:"category"`
	PostedDate      string           `json:"posted_date"`
	ImageUrl        string           `json:"image_url"`
	LastBuildDate   string           `json:"last_build_date"`
	PodcastEpisodes []PodcastEpisode `json:"podcast_episodes"`
	ArtistName      string           `json:"artist_name"`
	Explicit        string           `json:"explicit"`
}

type EpisodePlaybackHistory struct {
	YoutubeVideoId   string  `json:"youtube_video_id" gorm:"primary_key"`
	LastAccessDate   int64   `json:"last_access_date"`
	TotalTimeSkipped float64 `json:"total_time_skipped"`
}

func NewPodcastEpisodeFromPlaylist(youtubeVideo *youtube.PlaylistItem) PodcastEpisode {
	// For PlaylistItems, use VideoPublishedAt for the actual publication date
	publishedAt, err := time.Parse("2006-01-02T15:04:05Z07:00", youtubeVideo.ContentDetails.VideoPublishedAt)
	if err != nil {
		log.Error(err)
	}
	return PodcastEpisode{
		YoutubeVideoId:     youtubeVideo.Snippet.ResourceId.VideoId,
		EpisodeName:        youtubeVideo.Snippet.Title,
		EpisodeDescription: youtubeVideo.Snippet.Description,
		PublishedDate:      publishedAt,
		Type:               "PLAYLIST",
		PodcastId:          youtubeVideo.Snippet.PlaylistId,
	}
}

func NewPodcastEpisodeFromSearch(youtubeVideo *youtube.Video, duration time.Duration) PodcastEpisode {
	publishedAt, err := time.Parse("2006-01-02T15:04:05Z07:00", youtubeVideo.Snippet.PublishedAt)
	if err != nil {
		log.Error(err)
	}
	return PodcastEpisode{
		YoutubeVideoId:     youtubeVideo.Id,
		EpisodeName:        youtubeVideo.Snippet.Title,
		EpisodeDescription: youtubeVideo.Snippet.Description,
		PublishedDate:      publishedAt,
		Type:               "CHANNEL",
		PodcastId:          youtubeVideo.Snippet.ChannelId,
		Duration:           duration,
	}
}

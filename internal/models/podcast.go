package models

import (
	"ikoyhn/podcast-sponsorblock/internal/enum"
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
	ImageUrl           string        `json:"image_url"`
	Duration           time.Duration `json:"duration"`
}

type Podcast struct {
	AppleId         string           `json:"apple_id"`
	Id              string           `json:"id" gorm:"primary_key"`
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

func NewPodcastEpisode(youtubeVideo *youtube.Video, duration time.Duration, podcastType enum.PodcastType, podcastId string) PodcastEpisode {
	publishedAt, err := time.Parse("2006-01-02T15:04:05Z07:00", youtubeVideo.Snippet.PublishedAt)
	if err != nil {
		log.Error(err)
	}

	imageUrl := ""
	if youtubeVideo.Snippet.Thumbnails.Maxres != nil {
		imageUrl = youtubeVideo.Snippet.Thumbnails.Maxres.Url
	} else if youtubeVideo.Snippet.Thumbnails.Standard != nil {
		imageUrl = youtubeVideo.Snippet.Thumbnails.Standard.Url
	} else if youtubeVideo.Snippet.Thumbnails.High != nil {
		imageUrl = youtubeVideo.Snippet.Thumbnails.High.Url
	} else if youtubeVideo.Snippet.Thumbnails.Default != nil {
		imageUrl = youtubeVideo.Snippet.Thumbnails.Default.Url
	}

	return PodcastEpisode{
		YoutubeVideoId:     youtubeVideo.Id,
		EpisodeName:        youtubeVideo.Snippet.Title,
		EpisodeDescription: youtubeVideo.Snippet.Description,
		PublishedDate:      publishedAt,
		Type:               string(podcastType),
		PodcastId:          podcastId,
		Duration:           duration,
		ImageUrl:           imageUrl,
	}
}

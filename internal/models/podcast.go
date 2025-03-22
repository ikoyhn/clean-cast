package models

import (
	"google.golang.org/api/youtube/v3"
)

type PodcastEpisode struct {
	Id                 int32  `gorm:"autoIncrement;primary_key;not null"`
	YoutubeVideoId     string `json:"youtube_video_id" gorm:"index:youtubevideoid_type"`
	EpisodeName        string `json:"episode_name"`
	EpisodeDescription string `json:"episode_description"`
	PublishedDate      string `json:"published_date"`
	Type               string `json:"type" gorm:"index:youtubevideoid_type_channelid_type"`
	PodcastId          string `json:"podcast_id" gorm:"foreignkey:PodcastId;association_foreignkey:Id"`
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

func NewPodcastEpisode(youtubeVideo *youtube.PlaylistItem) PodcastEpisode {
	return PodcastEpisode{
		YoutubeVideoId:     youtubeVideo.Snippet.ResourceId.VideoId,
		EpisodeName:        youtubeVideo.Snippet.Title,
		EpisodeDescription: youtubeVideo.Snippet.Description,
		PublishedDate:      youtubeVideo.Snippet.PublishedAt,
		Type:               "PLAYLIST",
		PodcastId:          youtubeVideo.Snippet.PlaylistId,
	}
}

func NewPodcastEpisodeFromSearch(youtubeVideo *youtube.SearchResult) PodcastEpisode {
	return PodcastEpisode{
		YoutubeVideoId:     youtubeVideo.Id.VideoId,
		EpisodeName:        youtubeVideo.Snippet.Title,
		EpisodeDescription: youtubeVideo.Snippet.Description,
		PublishedDate:      youtubeVideo.Snippet.PublishedAt,
		Type:               "CHANNEL",
		PodcastId:          youtubeVideo.Snippet.ChannelId,
	}
}

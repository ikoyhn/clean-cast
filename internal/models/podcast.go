package models

type PodcastEpisode struct {
	YoutubeVideoId     string `json:"youtube_video_id" gorm:"primary_key"`
	EpisodeName        string `json:"episode_name"`
	EpisodeDescription string `json:"episode_description"`
	PublishedDate      string `json:"published_date"`
	Position           int64  `json:"position"`
}

type Podcast struct {
	AppleId          int              `json:"apple_id" gorm:"primary_key"`
	YoutubePodcastId string           `json:"youtube_podcast_id"`
	PodcastName      string           `json:"podcast_name"`
	Description      string           `json:"description"`
	Category         string           `json:"category"`
	Language         string           `json:"language"`
	PostedDate       string           `json:"posted_date"`
	ImageUrl         string           `json:"image_url"`
	LastBuildDate    string           `json:"last_build_date"`
	PodcastEpisodes  []PodcastEpisode `json:"podcast_episodes"`
	ArtistName       string           `json:"artist_name"`
	Explicit         string           `json:"explicit"`
}

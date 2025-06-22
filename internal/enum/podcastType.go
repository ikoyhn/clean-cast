package enum

type PodcastType string

const (
	PLAYLIST PodcastType = "PLAYLIST"
	CHANNEL  PodcastType = "CHANNEL"
)

type PodcastFetchType string

const (
	DATE    PodcastFetchType = "DATE"
	DEFAULT PodcastFetchType = "DEFAULT"
)

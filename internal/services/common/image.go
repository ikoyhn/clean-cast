package common

import (
	ytApi "google.golang.org/api/youtube/v3"
)

// SelectBestThumbnail selects the best available thumbnail from YouTube thumbnails
// Priority order: Maxres -> Standard -> High -> Default
func SelectBestThumbnail(thumbnails *ytApi.ThumbnailDetails) string {
	if thumbnails == nil {
		return ""
	}

	if thumbnails.Maxres != nil && thumbnails.Maxres.Url != "" {
		return thumbnails.Maxres.Url
	}

	if thumbnails.Standard != nil && thumbnails.Standard.Url != "" {
		return thumbnails.Standard.Url
	}

	if thumbnails.High != nil && thumbnails.High.Url != "" {
		return thumbnails.High.Url
	}

	if thumbnails.Default != nil && thumbnails.Default.Url != "" {
		return thumbnails.Default.Url
	}

	return ""
}

// SelectBestThumbnailFromSnippet is a convenience function that extracts thumbnails from a snippet
func SelectBestThumbnailFromSnippet(snippet *ytApi.ChannelSnippet) string {
	if snippet == nil {
		return ""
	}
	return SelectBestThumbnail(snippet.Thumbnails)
}

// SelectBestThumbnailFromVideoSnippet is a convenience function for video snippets
func SelectBestThumbnailFromVideoSnippet(snippet *ytApi.VideoSnippet) string {
	if snippet == nil {
		return ""
	}
	return SelectBestThumbnail(snippet.Thumbnails)
}

// SelectBestThumbnailFromPlaylistSnippet is a convenience function for playlist snippets
func SelectBestThumbnailFromPlaylistSnippet(snippet *ytApi.PlaylistSnippet) string {
	if snippet == nil {
		return ""
	}
	return SelectBestThumbnail(snippet.Thumbnails)
}

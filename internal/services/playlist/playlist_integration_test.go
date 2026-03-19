package playlist

import (
	"encoding/xml"
	"testing"

	"ikoyhn/podcast-sponsorblock/internal/services/generator"
	"ikoyhn/podcast-sponsorblock/internal/tests"
)

func TestGenerateRSSFromPlaylist_IT(t *testing.T) {
	tests.SetupIntegration(t)

	playlistID := "PLa7q8UDa6tvGRXE3-pdbiDQ-5_jRpqOkf"
	rssBytes := BuildPlaylistRssFeed(playlistID, "https://example.com")
	if len(rssBytes) == 0 {
		t.Fatalf("no RSS generated for playlist %s", playlistID)
	}

	var r struct {
		Channel generator.Podcast `xml:"channel"`
	}
	if err := xml.Unmarshal(rssBytes, &r); err != nil {
		t.Fatalf("generated RSS is not valid XML: %v", err)
	}
	if r.Channel.Title == "" {
		t.Fatalf("channel missing title")
	}
	if r.Channel.Description == "" {
		t.Fatalf("channel missing description")
	}
	if len(r.Channel.Items) == 0 {
		t.Fatalf("generated RSS contains no items")
	}
	it := r.Channel.Items[0]
	if it.Enclosure == nil || it.Enclosure.URL == "" {
		t.Fatalf("item missing enclosure url")
	}
	if it.IDuration == "" {
		t.Fatalf("item missing itunes:duration")
	}
	if it.IImage.Href == "" {
		t.Fatalf("item missing itunes:image")
	}
}

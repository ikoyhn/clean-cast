package models

import (
	"encoding/xml"
)

// OPML represents the root OPML document structure
type OPML struct {
	XMLName xml.Name   `xml:"opml"`
	Version string     `xml:"version,attr"`
	Head    OPMLHead   `xml:"head"`
	Body    OPMLBody   `xml:"body"`
}

// OPMLHead contains metadata about the OPML document
type OPMLHead struct {
	Title        string `xml:"title,omitempty"`
	DateCreated  string `xml:"dateCreated,omitempty"`
	DateModified string `xml:"dateModified,omitempty"`
	OwnerName    string `xml:"ownerName,omitempty"`
	OwnerEmail   string `xml:"ownerEmail,omitempty"`
}

// OPMLBody contains the outline structure
type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

// OPMLOutline represents a single outline (feed) entry
type OPMLOutline struct {
	Text    string `xml:"text,attr"`
	Title   string `xml:"title,attr,omitempty"`
	Type    string `xml:"type,attr"`
	XMLURL  string `xml:"xmlUrl,attr,omitempty"`
	HTMLURL string `xml:"htmlUrl,attr,omitempty"`
}

// OPMLImportRequest represents the request body for OPML import
type OPMLImportRequest struct {
	File []byte `json:"file"`
}

// OPMLImportResponse represents the response after OPML import
type OPMLImportResponse struct {
	Success      int      `json:"success"`
	Failed       int      `json:"failed"`
	TotalFeeds   int      `json:"total_feeds"`
	ErrorDetails []string `json:"error_details,omitempty"`
}

package generator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
)

const (
	pVersion = "1.3.1"
)

type Author struct {
	XMLName xml.Name `xml:"itunes:owner"`
	Name    string   `xml:"itunes:name"`
	Email   string   `xml:"itunes:email"`
}

// EnclosureType specifies the type of the enclosure.
const (
	M4A EnclosureType = iota
	M4V
	MP4
	MP3
	MOV
	PDF
	EPUB
)

const (
	enclosureDefault = "application/octet-stream"
)

// EnclosureType specifies the type of the enclosure.
type EnclosureType int

// String returns the MIME type encoding of the specified EnclosureType.
func (et EnclosureType) String() string {
	// https://help.apple.com/itc/podcasts_connect/#/itcb54353390
	switch et {
	case M4A:
		return "audio/mp4"
	case M4V:
		return "video/x-m4v"
	case MP4:
		return "video/mp4"
	case MP3:
		return "audio/mpeg"
	case MOV:
		return "video/quicktime"
	case PDF:
		return "application/pdf"
	case EPUB:
		return "document/x-epub"
	}
	return enclosureDefault
}

// Enclosure represents a download enclosure.
type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`

	// URL is the downloadable url for the content. (Required)
	URL string `xml:"url,attr"`

	// Length is the size in Bytes of the download. (Required)
	Length int64 `xml:"-"`
	// LengthFormatted is the size in Bytes of the download. (Required)
	//
	// This field gets overwritten with the API when setting Length.
	LengthFormatted string `xml:"length,attr"`

	// Type is MIME type encoding of the download. (Required)
	Type EnclosureType `xml:"-"`
	// TypeFormatted is MIME type encoding of the download. (Required)
	//
	// This field gets overwritten with the API when setting Type.
	TypeFormatted string `xml:"type,attr"`
}

// Image represents an image.
//
// Podcast feeds contain artwork that is a minimum size of
// 1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
// 72 dpi, in JPEG or PNG format with appropriate file
// extensions (.jpg, .png), and in the RGB colorspace. To optimize
// images for mobile devices, Apple recommends compressing your
// image files.
type Image struct {
	XMLName     xml.Name `xml:"image"`
	URL         string   `xml:"url"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description,omitempty"`
	Width       int      `xml:"width,omitempty"`
	Height      int      `xml:"height,omitempty"`
}

// Item represents a single entry in a podcast.
//
// Article minimal requirements are:
// - Title
// - Description
// - Link
//
// Audio minimal requirements are:
// - Title
// - Description
// - Enclosure (HREF, Type and Length all required)
//
// Recommendations:
// - Setting the minimal fields sets most of other fields, including iTunes.
// - Use the Published time.Time setting instead of PubDate.
// - Always set an Enclosure.Length, to be nice to your downloaders.
// - Use Enclosure.Type instead of setting TypeFormatted for valid extensions.
type Item struct {
	XMLName xml.Name `xml:"item"`
	GUID    struct {
		Value       string `xml:",chardata"`
		IsPermaLink bool   `xml:"isPermaLink,attr"`
	} `xml:"guid"`
	Title            string     `xml:"title"`
	Link             string     `xml:"link,omitempty"`
	Description      string     `xml:"description"`
	Author           string     `xml:"author,omitempty"`
	AuthorFormatted  string     `xml:"authorFormatted,omitempty"`
	Category         string     `xml:"category,omitempty"`
	Comments         string     `xml:"comments,omitempty"`
	Source           string     `xml:"source,omitempty"`
	PubDate          *time.Time `xml:"-"`
	PubDateFormatted string     `xml:"pubDate,omitempty"`
	Enclosure        *Enclosure

	// https://help.apple.com/itc/podcasts_connect/#/itcb54353390
	IAuthor            string `xml:"itunes:author,omitempty"`
	ISubtitle          string `xml:"itunes:subtitle,omitempty"`
	ISummary           *ISummary
	IDuration          string `xml:"itunes:duration,omitempty"`
	IExplicit          string `xml:"itunes:explicit,omitempty"`
	IIsClosedCaptioned string `xml:"itunes:isClosedCaptioned,omitempty"`
	IOrder             string `xml:"itunes:order,omitempty"`
}

// AddEnclosure adds the downloadable asset to the podcast Item.
func (i *Item) AddEnclosure(
	url string, enclosureType EnclosureType, lengthInBytes int64) {
	i.Enclosure = &Enclosure{
		URL:    url,
		Type:   enclosureType,
		Length: lengthInBytes,
	}
}

// AddPubDate adds the datetime as a parsed PubDate.
//
// UTC time is used by default.
func (i *Item) AddPubDate(datetime *time.Time) {
	i.PubDate = datetime
	i.PubDateFormatted = parseDateRFC1123Z(i.PubDate)
}

// AddSummary adds the iTunes summary.
//
// Limit: 4000 characters
//
// Note that this field is a CDATA encoded field which allows for rich text
// such as html links: `<a href="http://www.apple.com">Apple</a>`.
func (i *Item) AddSummary(summary string) {
	count := utf8.RuneCountInString(summary)
	if count > 4000 {
		s := []rune(summary)
		summary = string(s[0:4000])
	}
	i.ISummary = &ISummary{
		Text: summary,
	}
}

var parseDateRFC1123Z = func(t *time.Time) string {
	if t != nil && !t.IsZero() {
		return t.Format(time.RFC1123Z)
	}
	return time.Now().UTC().Format(time.RFC1123Z)
}

// Specifications: https://help.apple.com/itc/podcasts_connect/#/itcb54353390
//

// ICategory is a 2-tier classification system for iTunes.
type ICategory struct {
	XMLName     xml.Name `xml:"itunes:category"`
	Text        string   `xml:"text,attr"`
	ICategories []*ICategory
}

// ISummary is a 4000 character rich-text field for the itunes:summary tag.
//
// This is rendered as CDATA which allows for HTML tags such as `<a href="">`.
type ISummary struct {
	XMLName xml.Name `xml:"itunes:summary"`
	Text    string   `xml:",cdata"`
}

// Podcast represents a podcast.
type Podcast struct {
	XMLName        xml.Name `xml:"channel"`
	Title          string   `xml:"title"`
	Link           string   `xml:"link"`
	Description    string   `xml:"description"`
	Category       string   `xml:"category,omitempty"`
	Cloud          string   `xml:"cloud,omitempty"`
	Copyright      string   `xml:"copyright,omitempty"`
	Docs           string   `xml:"docs,omitempty"`
	Generator      string   `xml:"generator,omitempty"`
	Language       string   `xml:"language,omitempty"`
	LastBuildDate  string   `xml:"lastBuildDate,omitempty"`
	ManagingEditor string   `xml:"managingEditor,omitempty"`
	PubDate        string   `xml:"pubDate,omitempty"`
	Rating         string   `xml:"rating,omitempty"`
	SkipHours      string   `xml:"skipHours,omitempty"`
	SkipDays       string   `xml:"skipDays,omitempty"`
	TTL            int      `xml:"ttl,omitempty"`
	WebMaster      string   `xml:"webMaster,omitempty"`
	Image          *Image
	TextInput      *TextInput

	// https://help.apple.com/itc/podcasts_connect/#/itcb54353390
	IAuthor     string `xml:"itunes:author,omitempty"`
	ISubtitle   string `xml:"itunes:subtitle,omitempty"`
	ISummary    *ISummary
	IBlock      string  `xml:"itunes:block,omitempty"`
	IDuration   string  `xml:"itunes:duration,omitempty"`
	IExplicit   string  `xml:"itunes:explicit,omitempty"`
	IComplete   string  `xml:"itunes:complete,omitempty"`
	INewFeedURL string  `xml:"itunes:new-feed-url,omitempty"`
	IOwner      *Author // Author is formatted for itunes as-is
	ICategories []*ICategory

	Items []*Item

	encode func(w io.Writer, o interface{}) error
}

// TextInput represents text inputs.
type TextInput struct {
	XMLName     xml.Name `xml:"textInput"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Name        string   `xml:"name"`
	Link        string   `xml:"link"`
}

// New instantiates a Podcast with required parameters.
//
// Nil-able fields are optional but recommended as they are formatted
// to the expected proper formats.
func New(title, link, description string, lastBuildDate *time.Time) Podcast {
	return Podcast{
		Title:         title,
		Link:          link,
		Description:   description,
		Generator:     fmt.Sprintf("go podcast v%s (github.com/eduncan911/podcast)", pVersion),
		LastBuildDate: parseDateRFC1123Z(lastBuildDate),

		// setup dependency (could inject later)
		encode: encoder,
	}
}

// AddAuthor adds the specified Author to the podcast.
func (p *Podcast) AddAuthor(name, email string) {
	if len(email) == 0 {
		return
	}
	p.ManagingEditor = parseAuthorNameEmail(&Author{
		Name:  name,
		Email: email,
	})
	p.IAuthor = p.ManagingEditor
}

// AddCategory adds the category to the Podcast.
//
// ICategory can be listed multiple times.
//
// Calling this method multiple times will APPEND the category to the existing
// list, if any, including ICategory.
//
// Note that Apple iTunes has a specific list of categories that only can be
// used and will invalidate the feed if deviated from the list.  That list is
// as follows.
//
//   - Arts
//   - Design
//   - Fashion & Beauty
//   - Food
//   - Literature
//   - Performing Arts
//   - Visual Arts
//   - Business
//   - Business News
//   - Careers
//   - Investing
//   - Management & Marketing
//   - Shopping
//   - Comedy
//   - Education
//   - Education Technology
//   - Higher Education
//   - K-12
//   - Language Courses
//   - Training
//   - Games & Hobbies
//   - Automotive
//   - Aviation
//   - Hobbies
//   - Other Games
//   - Video Games
//   - Government & Organizations
//   - Local
//   - National
//   - Non-Profit
//   - Regional
//   - Health
//   - Alternative Health
//   - Fitness & Nutrition
//   - Self-Help
//   - Sexuality
//   - Kids & Family
//   - Music
//   - News & Politics
//   - Religion & Spirituality
//   - Buddhism
//   - Christianity
//   - Hinduism
//   - Islam
//   - Judaism
//   - Other
//   - Spirituality
//   - Science & Medicine
//   - Medicine
//   - Natural Sciences
//   - Social Sciences
//   - Society & Culture
//   - History
//   - Personal Journals
//   - Philosophy
//   - Places & Travel
//   - Sports & Recreation
//   - Amateur
//   - College & High School
//   - Outdoor
//   - Professional
//   - Technology
//   - Gadgets
//   - Podcasting
//   - Software How-To
//   - Tech News
//   - TV & Film
func (p *Podcast) AddCategory(category string, subCategories []string) {
	if len(category) == 0 {
		return
	}

	// RSS 2.0 Category only supports 1-tier
	if len(p.Category) > 0 {
		p.Category = p.Category + "," + category
	} else {
		p.Category = category
	}

	icat := ICategory{Text: category}
	for _, c := range subCategories {
		if len(c) == 0 {
			continue
		}
		icat2 := ICategory{Text: c}
		icat.ICategories = append(icat.ICategories, &icat2)
	}
	p.ICategories = append(p.ICategories, &icat)
}

// AddImage adds the specified Image to the Podcast.
//
// Podcast feeds contain artwork that is a minimum size of
// 1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
// 72 dpi, in JPEG or PNG format with appropriate file
// extensions (.jpg, .png), and in the RGB colorspace. To optimize
// images for mobile devices, Apple recommends compressing your
// image files.
func (p *Podcast) AddImage(url string) {
	if len(url) == 0 {
		return
	}
	p.Image = &Image{
		URL:   url,
		Title: p.Title,
		Link:  p.Link,
	}
}

// AddItem adds the podcast episode.  It returns a count of Items added or any
// errors in validation that may have occurred.
//
// This method takes the "itunes overrides" approach to populating
// itunes tags according to the overrides rules in the specification.
// This not only complies completely with iTunes parsing rules; but, it also
// displays what is possible to be set on an individual episode level – if you
// wish to have more fine grain control over your content.
//
// This method imposes strict validation of the Item being added to confirm
// to Podcast and iTunes specifications.
//
// Article minimal requirements are:
//
//   - Title
//   - Description
//   - Link
//
// Audio, Video and Downloads minimal requirements are:
//
//   - Title
//   - Description
//   - Enclosure (HREF, Type and Length all required)
//
// The following fields are always overwritten (don't set them):
//
//   - GUID
//   - PubDateFormatted
//   - AuthorFormatted
//   - Enclosure.TypeFormatted
//   - Enclosure.LengthFormatted
//
// Recommendations:
//
//   - Just set the minimal fields: the rest get set for you.
//   - Always set an Enclosure.Length, to be nice to your downloaders.
//   - Follow Apple's best practices to enrich your podcasts:
//     https://help.apple.com/itc/podcasts_connect/#/itc2b3780e76
//   - For specifications of itunes tags, see:
//     https://help.apple.com/itc/podcasts_connect/#/itcb54353390
func (p *Podcast) AddItem(i Item) (int, error) {
	// initial guards for required fields
	if len(i.Title) == 0 || len(i.Description) == 0 {
		return len(p.Items), errors.New("Title and Description are required")
	}
	if i.Enclosure != nil {
		if len(i.Enclosure.URL) == 0 {
			return len(p.Items),
				errors.New(i.Title + ": Enclosure.URL is required")
		}
		if i.Enclosure.Type.String() == enclosureDefault {
			return len(p.Items),
				errors.New(i.Title + ": Enclosure.Type is required")
		}
	} else if len(i.Link) == 0 {
		return len(p.Items),
			errors.New(i.Title + ": Link is required when not using Enclosure")
	}

	// corrective actions and overrides
	//
	i.PubDateFormatted = parseDateRFC1123Z(i.PubDate)
	if i.Enclosure != nil {

		if i.Enclosure.Length < 0 {
			i.Enclosure.Length = 0
		}
		i.Enclosure.LengthFormatted = strconv.FormatInt(i.Enclosure.Length, 10)
		i.Enclosure.TypeFormatted = i.Enclosure.Type.String()
	}

	// iTunes it
	//
	if len(i.IAuthor) == 0 {
		switch {
		case i.Author != "":
			i.IAuthor = i.Author
		}
	}

	p.Items = append(p.Items, &i)
	return len(p.Items), nil
}

// AddPubDate adds the datetime as a parsed PubDate.
//
// UTC time is used by default.
func (p *Podcast) AddPubDate(datetime *time.Time) {
	p.PubDate = parseDateRFC1123Z(datetime)
}

// AddLastBuildDate adds the datetime as a parsed PubDate.
//
// UTC time is used by default.
func (p *Podcast) AddLastBuildDate(datetime *time.Time) {
	p.LastBuildDate = parseDateRFC1123Z(datetime)
}

// AddSubTitle adds the iTunes subtitle that is displayed with the title
// in iTunes.
//
// Note that this field should be just a few words long according to Apple.
// This method will truncate the string to 64 chars if too long with "..."
func (p *Podcast) AddSubTitle(subTitle string) {
	count := utf8.RuneCountInString(subTitle)
	if count == 0 {
		return
	}
	if count > 64 {
		s := []rune(subTitle)
		subTitle = string(s[0:61]) + "..."
	}
	p.ISubtitle = subTitle
}

// AddSummary adds the iTunes summary.
//
// Limit: 4000 characters
//
// Note that this field is a CDATA encoded field which allows for rich text
// such as html links: `<a href="http://www.apple.com">Apple</a>`.
func (p *Podcast) AddSummary(summary string) {
	count := utf8.RuneCountInString(summary)
	if count == 0 {
		return
	}
	if count > 4000 {
		s := []rune(summary)
		summary = string(s[0:4000])
	}
	p.ISummary = &ISummary{
		Text: summary,
	}
}

// Bytes returns an encoded []byte slice.
func (p *Podcast) Bytes() []byte {
	return []byte(p.String())
}

// Encode writes the bytes to the io.Writer stream in RSS 2.0 specification.
func (p *Podcast) Encode(w io.Writer) error {
	if _, err := w.Write([]byte("<?xml version='1.0' encoding=\"UTF-8\"?>\n")); err != nil {
		return errors.Wrap(err, "podcast.Encode: w.Write return error")
	}

	wrapped := podcastWrapper{
		Version:   "2.0",
		ITUNESNS:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		ATOMNS:    "http://www.w3.org/2005/Atom",
		CONTENTNS: "http://purl.org/rss/1.0/modules/content/",
		Channel:   p,
	}
	return p.encode(w, wrapped)
}

// String encodes the Podcast state to a string.
func (p *Podcast) String() string {
	b := new(bytes.Buffer)
	if err := p.Encode(b); err != nil {
		return "String: podcast.write returned the error: " + err.Error()
	}
	return b.String()
}

// // Write implements the io.Writer interface to write an RSS 2.0 stream
// // that is compliant to the RSS 2.0 specification.
// func (p *Podcast) Write(b []byte) (n int, err error) {
// 	buf := bytes.NewBuffer(b)
// 	if err := p.Encode(buf); err != nil {
// 		return 0, errors.Wrap(err, "Write: podcast.encode returned error")
// 	}
// 	return buf.Len(), nil
// }

type podcastWrapper struct {
	XMLName   xml.Name `xml:"rss"`
	Version   string   `xml:"version,attr"`
	ATOMNS    string   `xml:"xmlns:atom,attr,omitempty"`
	ITUNESNS  string   `xml:"xmlns:itunes,attr"`
	CONTENTNS string   `xml:"xmlns:content,attr"`
	Channel   *Podcast
}

var encoder = func(w io.Writer, o interface{}) error {
	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	if err := e.Encode(o); err != nil {
		return errors.Wrap(err, "podcast.encoder: e.Encode returned error")
	}
	return nil
}

var parseAuthorNameEmail = func(a *Author) string {
	var author string
	if a != nil {
		author = a.Email
		if len(a.Name) > 0 {
			author = fmt.Sprintf("%s (%s)", a.Email, a.Name)
		}
	}
	return author
}

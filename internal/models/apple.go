package models

type AppleResponse struct {
	Feed Feed `json:"feed"`
}

type Feed struct {
	Author  Author  `json:"author"`
	Entries []Entry `json:"entry"`
	Updated string  `json:"updated"`
	Rights  string  `json:"rights"`
	Title   string  `json:"title"`
	Icon    string  `json:"icon"`
	Links   []Link  `json:"link"`
	ID      string  `json:"id"`
}

type Author struct {
	Name Name `json:"name"`
	URI  URI  `json:"uri"`
}

type Name struct {
	Label string `json:"label"`
}

type URI struct {
	Label string `json:"label"`
}

type Entry struct {
	Name        Name        `json:"im:name"`
	Images      []Image     `json:"im:image"`
	Summary     Summary     `json:"summary"`
	Price       Price       `json:"im:price"`
	ContentType ContentType `json:"im:contentType"`
	Rights      Rights      `json:"rights"`
	Title       Title       `json:"title"`
	Link        Link        `json:"link"`
	ID          ID          `json:"id"`
	Artist      Artist      `json:"im:artist"`
	Category    Category    `json:"category"`
	ReleaseDate ReleaseDate `json:"im:releaseDate"`
}

type Rights struct {
	Label string `json:"label"`
}

type Image struct {
	Label    string `json:"label"`
	Height   string `json:"height,attr"`
	Width    string `json:"width,attr"`
	URI      string `json:"uri,attr"`
	Alt      string `json:"alt,attr"`
	Length   string `json:"length,attr"`
	Type     string `json:"type,attr"`
	MimeType string `json:"mime-type,attr"`
}

type Summary struct {
	Label string `json:"label"`
}

type Price struct {
	Label    string `json:"label"`
	Amount   string `json:"amount,attr"`
	Currency string `json:"currency,attr"`
}

type ContentType struct {
	Term  string `json:"term,attr"`
	Label string `json:"label,attr"`
}

type Title struct {
	Label string `json:"label"`
}

type Link struct {
	Rel  string `json:"rel,attr"`
	Type string `json:"type,attr"`
	Href string `json:"href,attr"`
}

type ID struct {
	Label string `json:"label"`
	IMID  string `json:"im:id,attr"`
}

type Artist struct {
	Label string `json:"label"`
}

type Category struct {
	IMID   string `json:"im:id,attr"`
	Term   string `json:"term,attr"`
	Scheme string `json:"scheme,attr"`
	Label  string `json:"label,attr"`
}

type ReleaseDate struct {
	Label      string     `json:"label"`
	Attributes Attributes `json:"attributes"`
}

type Attributes struct {
	Label string `json:"label"`
}

type TopPodcast struct {
	Id          string
	Image       string
	Title       string
	Category    string
	Description string
}

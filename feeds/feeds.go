package feeds

type Image struct {
	ID      int    `json:"id"`
	Preview string `json:"preview"`
	Src     string `json:"src"`
}

type VideoFile struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Video struct {
	ID         int         `json:"id"`
	PreviewURL string      `json:"preview_url"`
	Files      []VideoFile `json:"files"`
}

type MediaFeed interface {
	FetchImage(query string, page, perPage int) ([]Image, int, error)
	FetchVideo(query string, page, perPage int) ([]Video, int, error)
	ImageDomain() string
	VideoDomain() string
}

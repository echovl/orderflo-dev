package pixabay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/layerhub-io/api/feeds"
)

const baseURL = "https://pixabay.com/api"

type imagesResponse struct {
	Total     int `json:"total"`
	TotalHits int `json:"total_hits"`
	Hits      []struct {
		ID           int    `json:"id"`
		WebFormatURL string `json:"webformatURL"`
		PreviewURL   string `json:"previewURL"`
	} `json:"hits"`
}

type videosResponse struct {
	Total     int        `json:"total"`
	TotalHits int        `json:"total_hits"`
	Hits      []videoHit `json:"hits"`
}

type video struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type videoHit struct {
	ID        int    `json:"id"`
	PictureID string `json:"picture_id"`
	Videos    struct {
		Large  video `json:"large"`
		Medium video `json:"medium"`
		Small  video `json:"small"`
		Tiny   video `json:"tiny"`
	} `json:"videos"`
}

type feed struct {
	apiKey string
}

func NewImageFeed(apiKey string) feeds.MediaFeed {
	return &feed{apiKey}
}

func (f *feed) ImageDomain() string {
	return "pixabay.com"
}

func (f *feed) VideoDomain() string {
	return "pixabay.com"
}

func (f *feed) FetchVideo(query string, page, perPage int) ([]feeds.Video, int, error) {
	encodedQuery := url.QueryEscape(query)

	if perPage < 3 {
		return nil, 0, fmt.Errorf("pixabay: perPage should not be less than 3")
	}

	url := fmt.Sprintf("%s/videos?key=%s&q=%s&image_type=photo&page=%v&per_page=%v", baseURL, f.apiKey, encodedQuery, page, perPage)

	rawResp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("pixabay: %w", err)
	}
	defer rawResp.Body.Close()

	resp := &videosResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return nil, 0, fmt.Errorf("pixabay: %w", err)
	}

	videos := make([]feeds.Video, len(resp.Hits))

	for i, hit := range resp.Hits {
		files := []feeds.VideoFile{
			{
				URL:    hit.Videos.Tiny.URL,
				Width:  hit.Videos.Tiny.Width,
				Height: hit.Videos.Tiny.Height,
			},
			{
				URL:    hit.Videos.Small.URL,
				Width:  hit.Videos.Small.Width,
				Height: hit.Videos.Small.Height,
			},
			{
				URL:    hit.Videos.Medium.URL,
				Width:  hit.Videos.Medium.Width,
				Height: hit.Videos.Medium.Height,
			},
		}

		videos[i] = feeds.Video{
			ID:         hit.ID,
			Files:      files,
			PreviewURL: fmt.Sprintf("https://i.vimeocdn.com/video/%s_640x360", hit.PictureID),
		}
	}

	return videos, resp.Total, nil
}

func (f *feed) FetchImage(query string, page, perPage int) ([]feeds.Image, int, error) {
	encodedQuery := url.QueryEscape(query)

	if perPage < 3 {
		return nil, 0, fmt.Errorf("pixabay: perPage should not be less than 3")
	}

	url := fmt.Sprintf("%s?key=%s&q=%s&image_type=photo&page=%v&per_page=%v", baseURL, f.apiKey, encodedQuery, page, perPage)

	rawResp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("pixabay: %w", err)
	}
	defer rawResp.Body.Close()

	resp := &imagesResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return nil, 0, fmt.Errorf("pixabay: %w", err)
	}

	images := make([]feeds.Image, len(resp.Hits))

	for i, hit := range resp.Hits {
		images[i] = feeds.Image{
			ID:      hit.ID,
			Src:     hit.WebFormatURL,
			Preview: hit.PreviewURL,
		}
	}

	return images, resp.Total, nil
}

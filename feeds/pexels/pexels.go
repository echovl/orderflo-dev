package pexels

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/layerhub-io/api/feeds"
)

const (
	baseImageURL = "https://api.pexels.com/v1/search"
	baseVideoURL = "https://api.pexels.com/videos/search"
)

type photo struct {
	ID  int `json:"id"`
	Src struct {
		Tiny     string `json:"tiny"`
		Original string `json:"original"`
		Small    string `json:"small"`
	} `json:"src"`
}

type videoFile struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Link   string `json:"link"`
}

type video struct {
	ID    int         `json:"id"`
	Image string      `json:"image"`
	Files []videoFile `json:"video_files"`
}

type imagesResponse struct {
	TotalResults int     `json:"total_results"`
	Photos       []photo `json:"photos"`
}

type videosResponse struct {
	TotalResults int     `json:"total_results"`
	Videos       []video `json:"videos"`
}

type feed struct {
	apiKey string
	client *http.Client
}

func NewImageFeed(apiKey string) feeds.MediaFeed {
	return &feed{apiKey, &http.Client{}}
}

func (f *feed) ImageDomain() string {
	return "images.pexels.com"
}

func (f *feed) VideoDomain() string {
	return "player.vimeo.com"
}

func (f *feed) FetchVideo(query string, page, perPage int) ([]feeds.Video, int, error) {
	encodedQuery := url.QueryEscape(query)

	url := fmt.Sprintf("%s?query=%s&page=%v&per_page=%v", baseVideoURL, encodedQuery, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}
	req.Header.Set("Authorization", f.apiKey)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}
	defer rawResp.Body.Close()

	resp := &videosResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}

	videos := make([]feeds.Video, len(resp.Videos))

	for i, hit := range resp.Videos {
		files := make([]feeds.VideoFile, len(hit.Files))
		for i, file := range hit.Files {
			files[i] = feeds.VideoFile{
				URL:    file.Link,
				Width:  file.Width,
				Height: file.Height,
			}
		}
		videos[i] = feeds.Video{
			ID:         hit.ID,
			PreviewURL: hit.Image,
			Files:      files,
		}
	}

	return videos, resp.TotalResults, nil
}

func (f *feed) FetchImage(query string, page, perPage int) ([]feeds.Image, int, error) {
	encodedQuery := url.QueryEscape(query)

	url := fmt.Sprintf("%s?query=%s&page=%v&per_page=%v", baseImageURL, encodedQuery, page, perPage)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}
	req.Header.Set("Authorization", f.apiKey)

	rawResp, err := f.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}
	defer rawResp.Body.Close()

	resp := &imagesResponse{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return nil, 0, fmt.Errorf("pexels: %w", err)
	}

	images := make([]feeds.Image, len(resp.Photos))

	for i, hit := range resp.Photos {
		images[i] = feeds.Image{
			ID:      hit.ID,
			Src:     hit.Src.Original,
			Preview: hit.Src.Small,
		}
	}

	return images, resp.TotalResults, nil
}

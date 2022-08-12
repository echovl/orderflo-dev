package layerhub

import (
	"context"

	"github.com/echovl/orderflo-dev/feeds"
)

func (c *Core) FetchPixabayVideos(ctx context.Context, query string, page, perPage int) ([]feeds.Video, int, error) {
	return c.pixabay.FetchVideo(query, page, perPage)
}

func (c *Core) FetchPexelsVideos(ctx context.Context, query string, page, perPage int) ([]feeds.Video, int, error) {
	return c.pexels.FetchVideo(query, page, perPage)
}

func (c *Core) FetchPixabayImages(ctx context.Context, query string, page, perPage int) ([]feeds.Image, int, error) {
	return c.pixabay.FetchImage(query, page, perPage)
}

func (c *Core) FetchPexelsImages(ctx context.Context, query string, page, perPage int) ([]feeds.Image, int, error) {
	return c.pexels.FetchImage(query, page, perPage)
}

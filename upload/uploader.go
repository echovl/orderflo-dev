package upload

import (
	"context"
)

type Uploader interface {
	Upload(ctx context.Context, key string, data []byte) (string, error)
	Download(ctx context.Context, key string) ([]byte, error)
}

type SignedUploader interface {
	Uploader
	GetPresignedURL(ctx context.Context, key string) (string, error)
}

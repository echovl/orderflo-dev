package layerhub

import (
	"context"
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/errors"
)

type Upload struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	ContentType string    `json:"content_type" db:"content_type"`
	Folder      string    `json:"folder" db:"folder"`
	Type        string    `json:"type" db:"type"`
	URL         string    `json:"url" db:"url"`
	CompanyID   string    `json:"company_id" db:"company_id"`
	CustomerID  string    `json:"customer_id" db:"customer_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

func NewUpload() *Upload {
	now := Now()
	return &Upload{
		ID:        UniqueID("upload"),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Core) FindUploads(ctx context.Context, filter *Filter) ([]Upload, int, error) {
	uploads, err := c.db.FindUploads(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := c.db.CountUploads(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return uploads, count, nil
}

func (c *Core) GetSignedURL(ctx context.Context, filename string) (string, error) {
	return c.uploader.GetPresignedURL(ctx, filename)
}

func (c *Core) PutUpload(ctx context.Context, upload *Upload) error {
	if err := c.db.PutUpload(ctx, upload); err != nil {
		return err
	}
	return nil
}

func (c *Core) GetUpload(ctx context.Context, id string) (*Upload, error) {
	uploads, err := c.db.FindUploads(ctx, &Filter{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(uploads) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("upload '%s' not found", id))
	}

	return &uploads[0], nil
}

func (c *Core) DeleteUpload(ctx context.Context, id string) error {
	return c.db.DeleteUpload(ctx, id)
}

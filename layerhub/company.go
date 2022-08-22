package layerhub

import (
	"context"
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/errors"
)

type Company struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewCompany() *Company {
	now := Now()
	return &Company{
		ID:        UniqueID("company"),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Core) PutCompany(ctx context.Context, company *Company) error {
	return c.db.PutCompany(ctx, company)
}

func (c *Core) GetCompany(ctx context.Context, id string) (*Company, error) {
	companies, err := c.db.FindCompanies(ctx, &Filter{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(companies) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("company '%s' not found", id))
	}

	return &companies[0], nil
}

func (c *Core) FindCompanies(ctx context.Context, filter *Filter) ([]Company, int, error) {
	companies, err := c.db.FindCompanies(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := c.db.CountCompanies(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return companies, count, nil
}

func (c *Core) DeleteCompany(ctx context.Context, id string) error {
	return c.db.DeleteCompany(ctx, id)
}

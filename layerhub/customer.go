package layerhub

import (
	"context"
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/errors"
)

type Customer struct {
	ID           string     `json:"id" db:"id"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CompanyID    string     `json:"company_id" db:"company_id"`
	Source       AuthSource `json:"source" db:"source"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

func NewCustomer() *Customer {
	now := Now()
	return &Customer{
		ID:        UniqueID("customer"),
		Source:    AuthSourceEmail,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Core) PutCustomer(ctx context.Context, customer *Customer) error {
	return c.db.PutCustomer(ctx, customer)
}

func (c *Core) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	customers, err := c.db.FindCustomers(ctx, &Filter{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	return &customers[0], nil
}

func (c *Core) FindCustomers(ctx context.Context, filter *Filter) ([]Customer, int, error) {
	customers, err := c.db.FindCustomers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := c.db.CountCustomers(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return customers, count, nil
}

func (c *Core) DeleteCustomer(ctx context.Context, id string) error {
	return c.db.DeleteCustomer(ctx, id)
}

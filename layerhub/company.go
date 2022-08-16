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
	UserID    string    `json:"user_id" db:"user_id"`
	ApiToken  string    `json:"api_token" db:"api_token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewCompany() *Company {
	now := Now()
	return &Company{
		ID:        UniqueID("company"),
		ApiToken:  NewApiToken(),
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
	return c.db.DeleteFont(ctx, id)
}

type Customer struct {
	ID        string    `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Email     string    `json:"email" db:"email"`
	CompanyID string    `json:"company_id" db:"company_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewCustomer() *Customer {
	now := Now()
	return &Customer{
		ID:        UniqueID("customer"),
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
	return c.db.DeleteFont(ctx, id)
}

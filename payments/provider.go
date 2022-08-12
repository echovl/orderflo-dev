package payments

import (
	"context"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type Billing struct {
	Interval string `json:"interval"`
	Price    string `json:"price"`
}

type Plan struct {
	ID                  string    `json:"id"`
	ProductID           string    `json:"product_id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Billing             []Billing `json:"billings"`
	AutoBillOutstanding bool      `json:"auto_bill_outstanding"`
	SetupFee            string    `json:"setup_fee"`
}

type Subscription struct {
	ID     string `json:"id"`
	PlanID string `json:"plan_id"`
	Status string `json:"status"`
}

type Provider interface {
	CreateProduct(ctx context.Context, product *Product) error

	GetProducts(ctx context.Context) ([]Product, error)

	CreatePlan(ctx context.Context, plan *Plan) error

	GetSubscription(ctx context.Context, id string) (*Subscription, error)
}

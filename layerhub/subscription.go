package layerhub

import (
	"context"

	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/payments"
)

type PaymentProvider string

const (
	Paypal PaymentProvider = "paypal"
	Stripe PaymentProvider = "stripe"
)

type Billing struct {
	ID                 string `json:"-" db:"id"`
	Interval           string `json:"interval" db:"interval"`
	Price              string `json:"price" db:"price"`
	SubscriptionPlanID string `json:"-" db:"subscription_plan_id"`
}

func NewBilling() *Billing {
	return &Billing{
		ID: UniqueID("billing"),
	}
}

type SubscriptionPlan struct {
	ID                  string          `json:"id" db:"id"`
	Provider            PaymentProvider `json:"provider" db:"provider"`
	ExternalID          string          `json:"external_id" db:"external_id"`
	Name                string          `json:"name" db:"name"`
	Description         string          `json:"description" db:"description"`
	ExternalProductID   string          `json:"external_product_id" db:"external_product_id"`
	Billing             []*Billing      `json:"billing"`
	AutoBillOutstanding bool            `json:"auto_bill_outstanding" db:"auto_bill_outstanding"`
	SetupFee            string          `json:"setup_fee" db:"setup_fee"`

	MaxTemplates int `json:"max_templates" bson:"max_templates" db:"max_templates"`
}

func NewSubscriptionPlan() *SubscriptionPlan {
	return &SubscriptionPlan{
		ID:       UniqueID("plan"),
		Provider: Paypal,
	}
}

// TODO: Verify planID using the paypal subscription
func (c *Core) SubscribeUser(ctx context.Context, userID, planID string) error {
	user := &User{}
	users, err := c.db.FindUsers(ctx, &Filter{ID: userID})
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return errors.E(errors.KindValidation, "user not found")
	}

	user.PlanID = planID
	// if err := c.storage.Put(ctx, user); err != nil {
	// 	return err
	// }

	return nil
}

func (c *Core) FindProducts(ctx context.Context) ([]payments.Product, error) {
	return c.paymentProvider.GetProducts(ctx)
}

func (c *Core) CreateProduct(ctx context.Context, product *payments.Product) error {
	return c.paymentProvider.CreateProduct(ctx, product)
}

func (c *Core) CreatePlan(ctx context.Context, plan *SubscriptionPlan) error {
	billing := make([]payments.Billing, len(plan.Billing))
	for i, b := range plan.Billing {
		billing[i] = payments.Billing{
			Interval: b.Interval,
			Price:    b.Price,
		}
	}

	providerPlan := &payments.Plan{
		Name:                plan.Name,
		Description:         plan.Description,
		ProductID:           string(plan.ExternalProductID),
		Billing:             billing,
		AutoBillOutstanding: plan.AutoBillOutstanding,
		SetupFee:            plan.SetupFee,
	}

	err := c.paymentProvider.CreatePlan(ctx, providerPlan)
	if err != nil {
		return err
	}

	plan.ExternalID = providerPlan.ID
	err = c.db.PutSubscriptionPlan(ctx, plan)
	if err != nil {
		return err
	}

	return nil
}

func (c *Core) ListPlan(ctx context.Context) ([]SubscriptionPlan, error) {
	return c.db.FindSubscriptionPlans(ctx)
}

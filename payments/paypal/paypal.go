package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/payments"
	"golang.org/x/oauth2/clientcredentials"
)

const baseURL = "https://api-m.sandbox.paypal.com/v1"

type oauth2Token struct {
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ExpiresAt   int64
}

type provider struct {
	oauth2Config *clientcredentials.Config
}

func NewPaymentProvider(clientID, secret string) payments.Provider {
	return &provider{
		oauth2Config: &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: secret,
			TokenURL:     baseURL + "/oauth2/token",
		},
	}
}

type paypalProduct struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
}

func (p *provider) GetProducts(ctx context.Context) ([]payments.Product, error) {
	resp := struct {
		Products []paypalProduct `json:"products"`
	}{}

	rawResp, err := p.DoRequest(ctx, http.MethodGet, baseURL+"/catalogs/products", nil)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}
	defer rawResp.Body.Close()

	err = json.NewDecoder(rawResp.Body).Decode(&resp)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}

	products := make([]payments.Product, len(resp.Products))
	for i, p := range resp.Products {
		products[i] = payments.Product{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	return products, nil
}

func (p *provider) CreateProduct(ctx context.Context, product *payments.Product) error {
	body, err := json.Marshal(paypalProduct{
		Name:        product.Name,
		Description: product.Description,
		Type:        "SERVICE",
		Category:    "SOFTWARE",
		ImageURL:    product.ImageURL,
	})
	if err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}

	rawResp, err := p.DoRequest(ctx, http.MethodPost, baseURL+"/catalogs/products", bytes.NewReader(body))
	if err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != http.StatusCreated {
		return errors.E(
			errors.KindUnexpected,
			errors.Errorf("paypal: %v", getResponseError(rawResp.Body)),
		)
	}

	resp := &paypalProduct{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}
	product.ID = resp.ID

	return nil
}

type frequency struct {
	IntervalUnit  string `json:"interval_unit"`
	IntervalCount int    `json:"interval_count"`
}

type money struct {
	Value        string `json:"value"`
	CurrencyCode string `json:"currency_code"`
}

type pricingScheme struct {
	FixedPrice money `json:"fixed_price"`
}

type billingCycle struct {
	Frequency     frequency     `json:"frequency"`
	TenureType    string        `json:"tenure_type"`
	Sequence      int           `json:"sequence"`
	TotalCycles   int           `json:"total_cycles"`
	PricingScheme pricingScheme `json:"pricing_scheme"`
}

type paymentPreferences struct {
	AutoBillOutstanding     bool   `json:"auto_bill_outstanding"`
	SetupFee                money  `json:"setup_fee"`
	SetupFeeFailureAction   string `json:"setup_fee_failure_action"`
	PaymentFailureThreshold int    `json:"payment_failure_threshold"`
}

type paypalPlan struct {
	ID                 string             `json:"id"`
	ProductID          string             `json:"product_id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	Status             string             `json:"status"`
	BillingCycles      []billingCycle     `json:"billing_cycles"`
	PaymentPreferences paymentPreferences `json:"payment_preferences"`
}

func (p *provider) CreatePlan(ctx context.Context, plan *payments.Plan) error {
	billing := make([]billingCycle, len(plan.Billing))
	for i, b := range plan.Billing {
		billing[i] = billingCycle{
			Frequency: frequency{
				IntervalUnit:  b.Interval,
				IntervalCount: 1,
			},
			TenureType:  "REGULAR",
			Sequence:    i + 1,
			TotalCycles: 0,
			PricingScheme: pricingScheme{
				FixedPrice: money{
					Value:        b.Price,
					CurrencyCode: "USD",
				},
			},
		}
	}

	body, err := json.Marshal(paypalPlan{
		ProductID:     plan.ProductID,
		Name:          plan.Name,
		Description:   plan.Description,
		Status:        "ACTIVE",
		BillingCycles: billing,
		PaymentPreferences: paymentPreferences{
			AutoBillOutstanding: plan.AutoBillOutstanding,
			SetupFee: money{
				CurrencyCode: "USD",
				Value:        plan.SetupFee,
			},
			SetupFeeFailureAction:   "CANCEL",
			PaymentFailureThreshold: 2,
		},
	})
	if err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}

	rawResp, err := p.DoRequest(ctx, http.MethodPost, baseURL+"/billing/plans", bytes.NewReader(body))
	if err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}
	defer rawResp.Body.Close()

	if rawResp.StatusCode != http.StatusCreated {
		return errors.E(
			errors.KindUnexpected,
			errors.Errorf("paypal: %v", getResponseError(rawResp.Body)),
		)
	}

	resp := &paypalPlan{}
	if err := json.NewDecoder(rawResp.Body).Decode(resp); err != nil {
		return errors.E(errors.KindUnexpected, errors.Errorf("paypal: %v", err))
	}

	plan.ID = resp.ID

	return nil
}

func (p *provider) GetSubscription(ctx context.Context, id string) (*payments.Subscription, error) {
	return nil, nil
}

func (p *provider) DoRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	client := p.oauth2Config.Client(ctx)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

type respError struct {
	Name    string        `json:"name"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

func getResponseError(body io.Reader) error {
	respErr := &respError{}

	if err := json.NewDecoder(body).Decode(respErr); err != nil {
		return err
	}

	return errors.Errorf(respErr.Message)
}

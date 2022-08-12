package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/layerhub"
	"github.com/layerhub-io/api/payments"
)

func (s *Server) handleSubscribeUser(c *fiber.Ctx) error {
	type request struct {
		SubscriptionID string `json:"subscription_id" validate:"required"`
	}

	type response struct {
		UserID         string `json:"user_id"`
		PlanID         string `json:"plan_id"`
		SubscriptionID string `json:"subscription_id"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.Validation(err)
	}

	planID := c.Params("id")
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return errors.Validation("user_id missing")
	}

	err := s.Core.SubscribeUser(c.Context(), userID, planID)
	if err != nil {
		return err
	}

	return c.JSON(response{userID, planID, req.SubscriptionID})

}

func (s *Server) handleListProducts(c *fiber.Ctx) error {
	type response struct {
		Products []payments.Product `json:"products"`
	}

	products, err := s.Core.FindProducts(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(response{products})
}

func (s *Server) handleCreateProduct(c *fiber.Ctx) error {
	type request struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
	}

	type response struct {
		Product *payments.Product `json:"product"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.Validation(err)
	}

	product := &payments.Product{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	err := s.Core.CreateProduct(c.Context(), product)
	if err != nil {
		return err
	}

	return c.JSON(response{product})
}

func (s *Server) handleListPlan(c *fiber.Ctx) error {
	type response struct {
		Plans []layerhub.SubscriptionPlan `json:"plans"`
	}

	plans, err := s.Core.ListPlan(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(response{plans})

}

func (s *Server) handleCreatePlan(c *fiber.Ctx) error {
	type billing struct {
		Interval string `json:"interval"`
		Price    string `json:"price"`
	}

	type request struct {
		ExternalProductID   string    `json:"external_product_id"`
		Name                string    `json:"name"`
		Description         string    `json:"description"`
		Billing             []billing `json:"billing"`
		AutoBillOutstanding bool      `json:"auto_bill_outstanding" validate:"required"`
		SetupFee            string    `json:"setup_fee" validate:"required"`
		MaxTemplates        int       `json:"max_templates"`
	}

	type response struct {
		Plan *layerhub.SubscriptionPlan `json:"plan"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.Validation(err)
	}

	cycles := make([]*layerhub.Billing, len(req.Billing))
	for i, b := range req.Billing {
		cycles[i] = layerhub.NewBilling()
		cycles[i].Price = b.Price
		cycles[i].Interval = b.Interval
	}

	plan := layerhub.NewSubscriptionPlan()
	plan.ExternalProductID = req.ExternalProductID
	plan.Name = req.Name
	plan.Description = req.Description
	plan.Billing = cycles
	plan.AutoBillOutstanding = req.AutoBillOutstanding
	plan.SetupFee = req.SetupFee
	plan.MaxTemplates = req.MaxTemplates

	err := s.Core.CreatePlan(c.Context(), plan)
	if err != nil {
		return err
	}

	return c.JSON(plan)
}

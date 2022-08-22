package http

import (
	"net/url"
	"time"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleListTemplate(c *fiber.Ctx) error {
	type request struct {
		CustomerID string `query:"customer_id"`
		CompanyID  string `query:"company_id"`
		Limit      int    `query:"limit"`
		Offset     int    `query:"offset"`
	}

	type response struct {
		Templates []layerhub.Template `json:"templates"`
		Total     int                 `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	filter := &layerhub.Filter{
		OptionalCustomerID: req.CustomerID,
		OptionalCompanyID:  session.Company.ID,
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	if session.Customer != nil {
		filter.OptionalCustomerID = session.Customer.ID
	}

	templates, count, err := s.Core.FindTemplates(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(response{templates, count})
}

func (s *Server) handleGetTemplate(c *fiber.Ctx) error {
	type response struct {
		Template *layerhub.Template `json:"template"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	template, err := s.Core.GetTemplate(c.Context(), id)
	if err != nil {
		return err
	}

	if !template.Public {
		if template.CompanyID != session.Company.ID {
			return errors.Authorization(template.ID)
		}

		if session.Customer != nil && template.CustomerID != session.Customer.ID {
			return errors.Authorization(template.ID)
		}
	}

	return c.JSON(response{template})
}

func (s *Server) handleRenderTemplate(c *fiber.Ctx) error {
	id := c.Params("id")

	query, err := url.ParseQuery(c.Context().QueryArgs().String())
	if err != nil {
		return err
	}

	template, err := s.Core.GetTemplate(c.Context(), id)
	if err != nil {
		return err
	}

	params := make(map[string]any)
	for k, v := range query {
		if len(v) == 1 {
			params[k] = v[0]
		}
	}

	img, err := s.Core.Render(c.Context(), template, params)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "image/png")

	return c.Send(img)
}

func (s *Server) handleCreateTemplate(c *fiber.Ctx) error {
	type request struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Layers      []*layerhub.Layer `json:"layers" validate:"required"`
		Tags        []string          `json:"tags"`
		Colors      []string          `json:"colors"`
		Frame       layerhub.Frame    `json:"frame" validate:"required"`
		Metadata    layerhub.Metadata `json:"metadata"`
		CustomerID  string            `json:"customer_id"`
		CompanyID   string            `json:"company_id"`
	}

	type response struct {
		Template *layerhub.Template `json:"template"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	template := layerhub.NewTemplate()
	template.Name = req.Name
	template.Description = req.Description
	template.Tags = req.Tags
	template.Colors = req.Colors
	template.Frame = req.Frame
	template.Metadata = req.Metadata
	template.Layers = req.Layers
	template.CompanyID = session.Company.ID

	if session.Customer != nil {
		template.CustomerID = req.CustomerID
	}

	if req.ID != "" {
		template.ID = req.ID
	}

	err := s.Core.PutTemplate(c.Context(), template)
	if err != nil {
		return err
	}

	return c.JSON(response{template})
}

func (s *Server) handleUpdateTemplate(c *fiber.Ctx) error {
	type request struct {
		Name        *string            `json:"name"`
		Description *string            `json:"description"`
		Tags        []string           `json:"tags"`
		Colors      []string           `json:"colors"`
		Layers      []*layerhub.Layer  `json:"layers"`
		Frame       *layerhub.Frame    `json:"frame"`
		Metadata    *layerhub.Metadata `json:"metadata"`
	}

	type response struct {
		Template *layerhub.Template `json:"template"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	template, err := s.Core.GetTemplate(c.Context(), id)
	if err != nil {
		return err
	}

	if template.CompanyID != session.Company.ID {
		return errors.Authorization(template.ID)
	}

	if session.Customer != nil && template.CustomerID != session.Customer.ID {
		return errors.Authorization(template.ID)
	}

	if err := assign.Structs(template, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	template.UpdatedAt = time.Now()

	err = s.Core.PutTemplate(c.Context(), template)
	if err != nil {
		return err
	}

	return c.JSON(response{template})
}

func (s *Server) handleDeleteTemplate(c *fiber.Ctx) error {
	type response struct {
		Template *layerhub.Template `json:"template"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	template, err := s.Core.GetTemplate(c.Context(), id)
	if err != nil {
		return err
	}

	if template.CompanyID != session.Company.ID {
		return errors.Authorization(template.ID)
	}

	if session.Customer != nil && template.CustomerID != session.Customer.ID {
		return errors.Authorization(template.ID)
	}

	err = s.Core.DeleteTemplate(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{template})
}

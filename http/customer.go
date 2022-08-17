package http

import (
	"fmt"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleCreateCustomer(c *fiber.Ctx) error {
	type request struct {
		FistName  string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		CompanyID string `json:"company_id"`
	}

	type response struct {
		Customer *layerhub.Customer `json:"customer"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	customer := layerhub.NewCustomer()
	customer.FirstName = req.FistName
	customer.LastName = req.LastName
	customer.Email = req.Email
	customer.UserID = session.UserID
	customer.CompanyID = session.CompanyID

	if session.IsWeb {
		if req.CompanyID == "" {
			return errors.E(errors.KindValidation, "'company_id' with value '' failed the 'required' validation")
		}

		company, err := s.Core.GetCompany(c.Context(), req.CompanyID)
		if err != nil {
			return err
		}

		if company.UserID != session.UserID {
			return errors.NotFound(fmt.Sprintf("company '%s' not found", company.ID))
		}

		customer.CompanyID = req.CompanyID
	}

	err := s.Core.PutCustomer(c.Context(), customer)
	if err != nil {
		return err
	}

	return c.JSON(response{customer})
}

func (s *Server) handleUpdateCustomer(c *fiber.Ctx) error {
	type request struct {
		FistName string `json:"first_name"`
		LastName string `json:"last_name"`
		Email    string `json:"email"`
	}

	type response struct {
		Customer *layerhub.Customer `json:"customer"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	customer, err := s.Core.GetCustomer(c.Context(), id)
	if err != nil {
		return err
	}

	if customer.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	if !session.IsWeb && (customer.CompanyID != session.CompanyID) {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	if err := assign.Structs(customer, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.Core.PutCustomer(c.Context(), customer)
	if err != nil {
		return err
	}

	return c.JSON(response{customer})
}

func (s *Server) handleGetCustomer(c *fiber.Ctx) error {
	type response struct {
		Customer *layerhub.Customer `json:"customer"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	customer, err := s.Core.GetCustomer(c.Context(), id)
	if errors.Is(err, errors.KindNotFound) {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	} else if err != nil {
		return err
	}

	if customer.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	if !session.IsWeb && (customer.CompanyID != session.CompanyID) {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	return c.JSON(response{customer})
}

func (s *Server) handleListCustomers(c *fiber.Ctx) error {
	type request struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}

	type response struct {
		Customers []layerhub.Customer `json:"customers"`
		Total     int                 `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	customers, count, err := s.Core.FindCustomers(c.Context(), &layerhub.Filter{
		CompanyID: session.CompanyID,
		UserID:    session.UserID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{Customers: customers, Total: count})
}

func (s *Server) handleDeleteCustomer(c *fiber.Ctx) error {
	type response struct {
		Customer *layerhub.Customer `json:"customer"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	customer, err := s.Core.GetCustomer(c.Context(), id)
	if err != nil {
		return err
	}

	if customer.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	if !session.IsWeb && (customer.CompanyID != session.CompanyID) {
		return errors.NotFound(fmt.Sprintf("customer '%s' not found", id))
	}

	err = s.Core.DeleteCustomer(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{customer})
}

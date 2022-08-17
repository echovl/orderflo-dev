package http

import (
	"fmt"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleCreateCompany(c *fiber.Ctx) error {
	type request struct {
		Name string `json:"name"`
	}

	type response struct {
		Company *layerhub.Company `json:"company"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	company := layerhub.NewCompany()
	company.Name = req.Name
	company.UserID = session.UserID

	err := s.Core.PutCompany(c.Context(), company)
	if err != nil {
		return err
	}

	return c.JSON(response{company})
}

func (s *Server) handleUpdateCompany(c *fiber.Ctx) error {
	type request struct {
		Name string `json:"name"`
	}

	type response struct {
		Company *layerhub.Company `json:"company"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	company, err := s.Core.GetCompany(c.Context(), id)
	if err != nil {
		return err
	}

	if company.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("company '%s' not found", id))
	}

	if err := assign.Structs(company, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.Core.PutCompany(c.Context(), company)
	if err != nil {
		return err
	}

	return c.JSON(response{company})
}

func (s *Server) handleGetCompany(c *fiber.Ctx) error {
	type response struct {
		Company *layerhub.Company `json:"company"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	company, err := s.Core.GetCompany(c.Context(), id)
	if errors.Is(err, errors.KindNotFound) {
		return errors.NotFound(fmt.Sprintf("company '%s' not found", id))
	} else if err != nil {
		return err
	}

	if company.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("company '%s' not found", id))
	}

	return c.JSON(response{company})
}

func (s *Server) handleListCompanies(c *fiber.Ctx) error {
	type request struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}

	type response struct {
		Companies []layerhub.Company `json:"companies"`
		Total     int                `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	companies, count, err := s.Core.FindCompanies(c.Context(), &layerhub.Filter{
		UserID: session.UserID,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{Companies: companies, Total: count})
}

func (s *Server) handleDeleteCompany(c *fiber.Ctx) error {
	type response struct {
		Company *layerhub.Company `json:"company"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	company, err := s.Core.GetCompany(c.Context(), id)
	if err != nil {
		return err
	}

	if company.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("company '%s' not found", id))
	}

	err = s.Core.DeleteCompany(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{company})
}

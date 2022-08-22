package http

import (
	"fmt"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

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

	if company.ID != session.Company.ID {
		return errors.NotFound(fmt.Sprintf("company '%s' not found", id))
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

	if company.ID != session.Company.ID {
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

package http

import (
	"fmt"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleEnableFonts(c *fiber.Ctx) error {
	type request struct {
		FontIDs    []string `json:"font_ids"`
		CustomerID string   `json:"customer_id" validate:"required"`
	}

	type response struct {
		FontIDs []string `json:"enabled_font_ids"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	err := s.Core.EnableFonts(c.Context(), req.CustomerID, req.FontIDs)
	if err != nil {
		return err
	}

	return c.JSON(response{req.FontIDs})
}

func (s *Server) handleDisableFonts(c *fiber.Ctx) error {
	type request struct {
		FontIDs    []string `json:"font_ids"`
		CustomerID string   `json:"customer_id" validate:"required"`
	}

	type response struct {
		FontIDs []string `json:"disable_font_ids"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	err := s.Core.DisableFonts(c.Context(), req.CustomerID, req.FontIDs)
	if err != nil {
		return err
	}

	return c.JSON(response{req.FontIDs})
}

func (s *Server) handleCreateFont(c *fiber.Ctx) error {
	type request struct {
		Family     string `json:"family"`
		FullName   string `json:"full_name"`
		Style      string `json:"style"`
		URL        string `json:"url"`
		Category   string `json:"category"`
		CustomerID string `json:"customer_id"`
		CompanyID  string `json:"company_id"`
	}

	type response struct {
		Font *layerhub.Font `json:"font"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	font := layerhub.NewFont()
	font.Family = req.Family
	font.FullName = req.FullName
	font.Style = req.Style
	font.URL = req.URL
	font.Category = req.Category
	font.CompanyID = session.Company.ID

	if session.Customer != nil {
		font.CustomerID = session.Customer.ID
	}

	err := s.Core.PutFont(c.Context(), font)
	if err != nil {
		return err
	}

	return c.JSON(response{font})
}

func (s *Server) handleUpdateFont(c *fiber.Ctx) error {
	type request struct {
		Family   *string `json:"family"`
		FullName *string `json:"full_name"`
		Style    *string `json:"style"`
		URL      *string `json:"url"`
	}

	type response struct {
		Font *layerhub.Font `json:"font"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	font, err := s.Core.GetFont(c.Context(), id)
	if err != nil {
		return err
	}

	if font.CompanyID != session.Company.ID {
		return errors.Authorization(font.ID)
	}

	if session.Customer != nil && font.CustomerID != session.Customer.ID {
		return errors.Authorization(font.ID)
	}

	if err := assign.Structs(font, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.Core.PutFont(c.Context(), font)
	if err != nil {
		return err
	}

	return c.JSON(response{font})
}

func (s *Server) handleGetFont(c *fiber.Ctx) error {
	type response struct {
		Font *layerhub.Font `json:"font"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	font, err := s.Core.GetFont(c.Context(), id)
	if errors.Is(err, errors.KindNotFound) {
		return errors.NotFound(fmt.Sprintf("font '%s' not found", id))
	} else if err != nil {
		return err
	}

	if !font.Public {
		if font.CompanyID != session.Company.ID {
			return errors.Authorization(font.ID)
		}

		if session.Customer != nil && font.CustomerID != session.Customer.ID {
			return errors.Authorization(font.ID)
		}
	}

	return c.JSON(response{font})
}

func (s *Server) handleListFonts(c *fiber.Ctx) error {
	type request struct {
		CustomerID     string `query:"customer_id"`
		PostscriptName string `query:"postscript_name"`
		Enabled        *bool  `query:"enabled"`
		Limit          int    `query:"limit"`
		Offset         int    `query:"offset"`
	}

	type response struct {
		Fonts []layerhub.Font `json:"fonts"`
		Total int             `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	filter := &layerhub.Filter{
		OptionalCustomerID: req.CustomerID,
		OptionalCompanyID:  session.Company.ID,
		PostscriptName:     req.PostscriptName,
		EnabledFonts:       req.Enabled,
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	if session.Customer != nil {
		filter.OptionalCustomerID = session.Customer.ID
	}

	fonts, count, err := s.Core.FindFonts(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(response{Fonts: fonts, Total: count})
}

func (s *Server) handleDeleteFont(c *fiber.Ctx) error {
	type response struct {
		Font *layerhub.Font `json:"font"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))
	font, err := s.Core.GetFont(c.Context(), id)
	if err != nil {
		return err
	}

	if font.CompanyID != session.Company.ID {
		return errors.Authorization(font.ID)
	}

	if session.Customer != nil && font.CustomerID != session.Customer.ID {
		return errors.Authorization(font.ID)
	}

	err = s.Core.DeleteFont(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{font})
}

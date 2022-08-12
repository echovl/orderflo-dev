package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/layerhub-io/api/assign"
	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/layerhub"
)

func (s *Server) handleEnableFonts(c *fiber.Ctx) error {
	type request struct {
		FontIDs []string `json:"font_ids"`
	}

	type response struct {
		FontIDs []string `json:"enabled_font_ids"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)

	err := s.Core.EnableFonts(c.Context(), session.UserID, req.FontIDs)
	if err != nil {
		return err
	}

	return c.JSON(response{req.FontIDs})
}

func (s *Server) handleDisableFonts(c *fiber.Ctx) error {
	type request struct {
		FontIDs []string `json:"font_ids"`
	}

	type response struct {
		FontIDs []string `json:"disable_font_ids"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)

	err := s.Core.DisableFonts(c.Context(), session.UserID, req.FontIDs)
	if err != nil {
		return err
	}

	return c.JSON(response{req.FontIDs})
}

func (s *Server) handleCreateFont(c *fiber.Ctx) error {
	type request struct {
		Family   string `json:"family"`
		FullName string `json:"full_name"`
		Style    string `json:"style"`
		URL      string `json:"url"`
		Category string `json:"category"`
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

	if !session.IsAdmin() {
		font.UserID = session.UserID
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

	if !session.IsAdmin() && font.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("font '%s' not found", id))
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

	if !session.IsAdmin() && font.UserID != session.UserID && font.UserID != "" {
		return errors.NotFound(fmt.Sprintf("font '%s' not found", id))
	}

	return c.JSON(response{font})
}

func (s *Server) handleListFonts(c *fiber.Ctx) error {
	type request struct {
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
	fonts, count, err := s.Core.FindFonts(c.Context(), &layerhub.Filter{
		UserID:         session.UserID,
		Limit:          req.Limit,
		Offset:         req.Offset,
		PostscriptName: req.PostscriptName,
		FontEnabled:    req.Enabled,
	})
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

	if !session.IsAdmin() && font.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("font '%s' not found", id))
	}

	err = s.Core.DeleteFont(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{font})
}

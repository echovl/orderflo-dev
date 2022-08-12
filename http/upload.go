package http

import (
	"fmt"
	"mime"
	"path"

	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
)

func (s *Server) handleListUpload(c *fiber.Ctx) error {
	type request struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}

	type response struct {
		Uploads []layerhub.Upload `json:"uploads"`
		Total   int               `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	uploads, count, err := s.Core.FindUploads(c.Context(), &layerhub.Filter{
		UserID: session.UserID,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{uploads, count})
}

func (s *Server) handleCreateSignedURL(c *fiber.Ctx) error {
	type request struct {
		Filename string `json:"filename" validate:"required"`
	}

	type response struct {
		URL string `json:"url"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	if mime.TypeByExtension(path.Ext(req.Filename)) == "" {
		return errors.E(errors.KindValidation, "unknown filename extension")
	}

	url, err := s.Core.GetSignedURL(c.Context(), req.Filename)
	if err != nil {
		return err
	}

	return c.JSON(response{url})
}

func (s *Server) handleCreateUpload(c *fiber.Ctx) error {
	type request struct {
		Filename string `json:"filename" validate:"required"`
	}

	type response struct {
		Upload *layerhub.Upload `json:"upload"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	upload, err := s.Core.CreateUpload(
		c.Context(),
		session.UserID,
		req.Filename)
	if err != nil {
		return err
	}

	return c.JSON(response{upload})
}

func (s *Server) handleDeleteUpload(c *fiber.Ctx) error {
	type response struct {
		Upload *layerhub.Upload `json:"upload"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")

	upload, err := s.Core.GetUpload(c.Context(), id)
	if err != nil {
		return err
	}

	if !session.IsAdmin() && upload.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("upload '%s' not found", id))
	}

	err = s.Core.DeleteUpload(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{upload})
}

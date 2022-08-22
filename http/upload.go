package http

import (
	"mime"
	"path"

	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleListUpload(c *fiber.Ctx) error {
	type request struct {
		CustomerID string `query:"customer_id"`
		Limit      int    `query:"limit"`
		Offset     int    `query:"offset"`
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
	filter := &layerhub.Filter{
		OptionalCustomerID: req.CustomerID,
		OptionalCompanyID:  session.Company.ID,
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	if session.Customer != nil {
		filter.OptionalCustomerID = session.Customer.ID
	}

	uploads, count, err := s.Core.FindUploads(c.Context(), filter)
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
	upload := layerhub.NewUpload()
	upload.CompanyID = session.Company.ID

	if session.Customer != nil {
		upload.CustomerID = session.Customer.ID
	}

	err := s.Core.PutUpload(c.Context(), upload)
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

	if upload.CompanyID != session.Company.ID {
		return errors.Authorization(upload.ID)
	}

	if session.Customer != nil && upload.CustomerID != session.Customer.ID {
		return errors.Authorization(upload.ID)
	}

	err = s.Core.DeleteUpload(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{upload})
}

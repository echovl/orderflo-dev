package http

import (
	"github.com/aws/smithy-go/ptr"
	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleCreateFrame(c *fiber.Ctx) error {
	type request struct {
		Name       string             `json:"name"`
		Visibility string             `json:"visibility"`
		Width      float64            `json:"width"`
		Height     float64            `json:"height"`
		Unit       layerhub.FrameUnit `json:"unit" validate:"oneof=cm px in"`
		CustomerID string             `json:"customer_id"`
		CompanyID  string             `json:"company_id"`
	}

	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	frame := layerhub.NewFrame()
	frame.Name = req.Name
	frame.Width = req.Width
	frame.Height = req.Height
	frame.Unit = req.Unit
	frame.UserID = session.UserID

	if !session.IsWeb {
		if req.CustomerID == "" {
			return errors.Validation("'request.customer_id' with value '' failed the 'required' validation")
		}
		frame.CustomerID = req.CustomerID
		frame.CompanyID = session.CompanyID
	} else {
		frame.CompanyID = req.CompanyID
	}

	err := s.Core.PutFrame(c.Context(), frame)
	if err != nil {
		return err
	}

	return c.JSON(response{frame})
}

func (s *Server) handleUpdateFrame(c *fiber.Ctx) error {
	type request struct {
		Name   *string             `json:"name"`
		Width  *float64            `json:"width"`
		Height *float64            `json:"height"`
		Unit   *layerhub.FrameUnit `json:"unit"`
	}

	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if frame.UserID != session.UserID {
		return errors.Authorization(frame.ID)
	}

	if !session.IsWeb && frame.CompanyID != session.CompanyID {
		return errors.Authorization(frame.ID)
	}

	if err := assign.Structs(frame, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.Core.PutFrame(c.Context(), frame)
	if err != nil {
		return err
	}

	return c.JSON(response{frame})
}

func (s *Server) handleGetFrame(c *fiber.Ctx) error {
	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if !frame.Public {
		if frame.UserID != session.UserID {
			return errors.Authorization(frame.ID)
		}

		if !session.IsWeb && frame.CompanyID != "" && frame.CompanyID != session.CompanyID {
			return errors.Authorization(frame.ID)
		}
	}

	return c.JSON(response{frame})
}

func (s *Server) handleListFrames(c *fiber.Ctx) error {
	type request struct {
		CustomerID string `query:"customer_id"`
		CompanyID  string `query:"company_id"`
		Limit      int    `query:"limit"`
		Offset     int    `query:"offset"`
	}

	type response struct {
		Frames []layerhub.Frame `json:"frames"`
		Total  int              `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	filter := &layerhub.Filter{
		OptionalCustomerID: req.CustomerID,
		OptionalCompanyID:  session.CompanyID,
		OptionalUserID:     session.UserID,
		UsedInTemplate:     ptr.Bool(false),
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	if session.IsWeb {
		filter.OptionalCompanyID = req.CompanyID
	}

	frames, count, err := s.Core.FindFrames(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(response{frames, count})
}
func (s *Server) handleDeleteFrame(c *fiber.Ctx) error {
	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if frame.UserID != session.UserID {
		return errors.Authorization(frame.ID)
	}

	if !session.IsWeb && frame.CompanyID != session.CompanyID {
		return errors.Authorization(frame.ID)
	}

	err = s.Core.DeleteFrame(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{frame})
}

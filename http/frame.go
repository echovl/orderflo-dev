package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
)

func (s *Server) handleCreateFrame(c *fiber.Ctx) error {
	type request struct {
		Name       string             `json:"name"`
		Visibility string             `json:"visibility"`
		Width      float64            `json:"width"`
		Height     float64            `json:"height"`
		Unit       layerhub.FrameUnit `json:"unit" validate:"oneof=cm px in"`
	}

	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	frame := layerhub.NewFrame()
	frame.Name = req.Name
	frame.Visibility = layerhub.FramePublic
	frame.Width = req.Width
	frame.Height = req.Height
	frame.Unit = req.Unit

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

	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if frame.Visibility != layerhub.FramePublic {
		return errors.NotFound(fmt.Sprintf("frame '%s' not found", id))
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

	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if frame.Visibility != layerhub.FramePublic {
		return errors.NotFound(fmt.Sprintf("frame '%s' not found", id))
	}

	return c.JSON(response{frame})
}

func (s *Server) handleListFrames(c *fiber.Ctx) error {
	type request struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}

	type response struct {
		Frames []layerhub.Frame `json:"frames"`
		Total  int              `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	frames, count, err := s.Core.FindFrames(c.Context(), &layerhub.Filter{
		Visibility: layerhub.FramePublic,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{frames, count})
}
func (s *Server) handleDeleteFrame(c *fiber.Ctx) error {
	type response struct {
		Frame *layerhub.Frame `json:"frame"`
	}

	id := c.Params("id")
	frame, err := s.Core.GetFrame(c.Context(), id)
	if err != nil {
		return err
	}

	if frame.Visibility != layerhub.FramePublic {
		return errors.NotFound(fmt.Sprintf("frame '%s' not found", id))
	}

	err = s.Core.DeleteFrame(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{frame})
}

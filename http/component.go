package http

import (
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleCreateComponent(c *fiber.Ctx) error {
	type request struct {
		Name       string            `json:"name"`
		Layers     []*layerhub.Layer `json:"layers" validate:"required"`
		Metadata   map[string]any    `json:"metadata"`
		CustomerID string            `json:"customer_id"`
	}

	type response struct {
		Component *layerhub.Component `json:"component"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	component := layerhub.NewComponent()
	component.Name = req.Name
	component.Layers = req.Layers
	component.Metadata = req.Metadata
	component.UserID = session.UserID

	if !session.IsWeb {
		if req.CustomerID == "" {
			return errors.E(errors.KindValidation, "'customer_id' with value '' failed the 'required' validation")
		}
		component.CustomerID = req.CustomerID
		component.CompanyID = session.CompanyID
	} else {
		component.Public = true
	}

	err := s.Core.PutComponent(c.Context(), component)
	if err != nil {
		return err
	}

	return c.JSON(response{component})
}

func (s *Server) handleUpdateComponent(c *fiber.Ctx) error {
	type request struct {
		Name     *string           `json:"name"`
		Layers   []*layerhub.Layer `json:"layers"`
		Metadata map[string]any    `json:"metadata"`
	}

	type response struct {
		Component *layerhub.Component `json:"component"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)

	id := c.Params("id")
	component, err := s.Core.GetComponent(c.Context(), id)
	if err != nil {
		return err
	}

	if !session.IsAdmin() && component.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("component '%s' not found", id))
	}

	if err := assign.Structs(component, req); err != nil {
		return err
	}
	component.UpdatedAt = time.Now()

	err = s.Core.PutComponent(c.Context(), component)
	if err != nil {
		return err
	}

	return c.JSON(response{component})
}

func (s *Server) handleGetComponent(c *fiber.Ctx) error {
	type response struct {
		Component *layerhub.Component `json:"component"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")

	component, err := s.Core.GetComponent(c.Context(), id)
	if err != nil {
		return err
	}

	if !component.Public {
		if component.UserID != session.UserID {
			return errors.NotFound(fmt.Sprintf("component '%s' not found", id))
		}

		if !session.IsWeb && component.CompanyID != session.CompanyID {
			return errors.NotFound(fmt.Sprintf("component '%s' not found", id))
		}
	}

	return c.JSON(response{component})
}

func (s *Server) handleListComponent(c *fiber.Ctx) error {
	type request struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}

	type response struct {
		Components []layerhub.Component `json:"components"`
		Total      int                  `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	components, count, err := s.Core.FindComponents(c.Context(), &layerhub.Filter{
		UserID: session.UserID,
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{components, count})
}

func (s *Server) handleDeleteComponent(c *fiber.Ctx) error {
	type response struct {
		Component *layerhub.Component `json:"component"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")

	component, err := s.Core.GetComponent(c.Context(), id)
	if err != nil {
		return err
	}

	if !session.IsAdmin() && component.UserID != session.UserID {
		return errors.NotFound(fmt.Sprintf("component '%s' not found", id))
	}

	err = s.Core.DeleteComponent(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{component})
}

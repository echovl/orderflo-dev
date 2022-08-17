package http

import (
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) handleListProject(c *fiber.Ctx) error {
	type request struct {
		CustomerID string `query:"customer_id"`
		Limit      int    `query:"limit"`
		Offset     int    `query:"offset"`
	}

	type response struct {
		Projects []layerhub.Project `json:"projects"`
		Total    int                `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	projects, count, err := s.Core.FindProjects(c.Context(), &layerhub.Filter{
		CustomerID: req.CustomerID,
		CompanyID:  session.CompanyID,
		UserID:     session.UserID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return err
	}

	return c.JSON(response{projects, count})
}

func (s *Server) handleGetProject(c *fiber.Ctx) error {
	type response struct {
		Project *layerhub.Project `json:"project"`
	}

	session, _ := s.getSession(c)
	id := c.Params("id")

	project, err := s.Core.GetProject(c.Context(), id)
	if err != nil {
		return err
	}

	if project.UserID != session.UserID || project.CompanyID != session.CompanyID {
		return errors.NotFound(fmt.Sprintf("project '%s' not found", id))
	}

	return c.JSON(response{project})
}

func (s *Server) handleCreateProject(c *fiber.Ctx) error {
	type request struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Layers      []*layerhub.Layer `json:"layers" validate:"required"`
		Frame       layerhub.Frame    `json:"frame" validate:"required"`
		CustomerID  string            `json:"customer_id" validate:"required"`
	}

	type response struct {
		Project *layerhub.Project `json:"project"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)

	project := layerhub.NewProject()
	project.Name = req.Name
	project.Description = req.Description
	project.Frame = req.Frame
	project.Layers = req.Layers
	project.CustomerID = req.CustomerID
	project.CompanyID = session.CompanyID
	project.UserID = session.UserID
	if req.ID != "" {
		project.ID = req.ID
	}

	err := s.Core.PutProject(c.Context(), project)
	if err != nil {
		return err
	}

	return c.JSON(response{project})
}

func (s *Server) handleUpdateProject(c *fiber.Ctx) error {
	type request struct {
		Name        *string           `json:"name"`
		Description *string           `json:"description"`
		Layers      []*layerhub.Layer `json:"layers"`
		Frame       *layerhub.Frame   `json:"frame"`
	}

	type response struct {
		Project *layerhub.Project `json:"project"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))

	project, err := s.Core.GetProject(c.Context(), id)
	if err != nil {
		return err
	}

	if project.UserID != session.UserID || project.CompanyID != session.CompanyID {
		return errors.NotFound(fmt.Sprintf("project '%s' not found", id))
	}

	if err := assign.Structs(project, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	project.UpdatedAt = time.Now()

	err = s.Core.PutProject(c.Context(), project)
	if err != nil {
		return err
	}

	return c.JSON(response{project})
}

func (s *Server) handleDeleteProject(c *fiber.Ctx) error {
	type response struct {
		Project *layerhub.Project `json:"project"`
	}

	session, _ := s.getSession(c)
	id := string(c.Params("id"))

	project, err := s.Core.GetProject(c.Context(), id)
	if err != nil {
		return err
	}

	if project.UserID != session.UserID || project.CompanyID != session.CompanyID {
		return errors.NotFound(fmt.Sprintf("project '%s' not found", id))
	}

	err = s.Core.DeleteProject(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(response{project})
}

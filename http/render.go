package http

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
)

func (s *Server) handleRenderDesign(c *fiber.Ctx) error {
	id := c.Params("id")

	query, err := url.ParseQuery(c.Context().QueryArgs().String())
	if err != nil {
		return err
	}

	filter := &layerhub.Filter{RegularOrShortID: id, Limit: 1}
	templates, _, err := s.Core.FindTemplates(c.Context(), filter)
	if err != nil {
		return err
	}

	projects, _, err := s.Core.FindProjects(c.Context(), filter)
	if err != nil {
		return err
	}

	params := make(map[string]any)
	for k, v := range query {
		if len(v) == 1 {
			params[k] = v[0]
		}
	}

	var schema any
	if len(templates) == 1 {
		schema, err = s.Core.GetTemplate(c.Context(), id)
		if err != nil {
			return err
		}
	} else if len(projects) == 1 {
		schema, err = s.Core.GetProject(c.Context(), id)
		if err != nil {
			return err
		}
	} else {
		return errors.NotFound(fmt.Sprintf("template or project '%s' not found", id))
	}

	img, err := s.Core.Render(c.Context(), schema, params)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "image/png")

	return c.Send(img)
}

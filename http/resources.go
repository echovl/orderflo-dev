package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/feeds"
)

func (s *Server) handleFetchPixabayVideos(c *fiber.Ctx) error {
	type request struct {
		Query   string `query:"query"`
		Page    int    `query:"page" validate:"required,min=0"`
		PerPage int    `query:"per_page" validate:"required,min=1"`
	}

	type response struct {
		Videos  []feeds.Video `json:"videos"`
		Page    int           `json:"page"`
		PerPage int           `json:"per_page"`
		Total   int           `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	videos, total, err := s.Core.FetchPixabayVideos(
		c.Context(),
		req.Query,
		req.Page,
		req.PerPage,
	)
	if err != nil {
		return err
	}

	return c.JSON(response{
		Videos:  videos,
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   total,
	})
}

func (s *Server) handleFetchPixabayImages(c *fiber.Ctx) error {
	type request struct {
		Query   string `query:"query"`
		Page    int    `query:"page" validate:"required,min=0"`
		PerPage int    `query:"per_page" validate:"required,min=1"`
	}

	type response struct {
		Images  []feeds.Image `json:"images"`
		Page    int           `json:"page"`
		PerPage int           `json:"per_page"`
		Total   int           `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	images, total, err := s.Core.FetchPixabayImages(
		c.Context(),
		req.Query,
		req.Page,
		req.PerPage,
	)
	if err != nil {
		return err
	}

	return c.JSON(response{
		Images:  images,
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   total,
	})
}

func (s *Server) handleFetchPexelsVideos(c *fiber.Ctx) error {
	type request struct {
		Query   string `query:"query"`
		Page    int    `query:"page" validate:"required,min=0"`
		PerPage int    `query:"per_page" validate:"required,min=1"`
	}

	type response struct {
		Videos  []feeds.Video `json:"videos"`
		Page    int           `json:"page"`
		PerPage int           `json:"per_page"`
		Total   int           `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	videos, total, err := s.Core.FetchPexelsVideos(
		c.Context(),
		req.Query,
		req.Page,
		req.PerPage,
	)
	if err != nil {
		return err
	}

	return c.JSON(response{
		Videos:  videos,
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   total,
	})
}

func (s *Server) handleFetchPexelsImages(c *fiber.Ctx) error {
	type request struct {
		Query   string `query:"query"`
		Page    int    `query:"page" validate:"required,min=0"`
		PerPage int    `query:"per_page" validate:"required,min=1"`
	}

	type response struct {
		Images  []feeds.Image `json:"images"`
		Page    int           `json:"page"`
		PerPage int           `json:"per_page"`
		Total   int           `json:"total"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	images, total, err := s.Core.FetchPexelsImages(
		c.Context(),
		req.Query,
		req.Page,
		req.PerPage,
	)
	if err != nil {
		return err
	}

	return c.JSON(response{
		Images:  images,
		Page:    req.Page,
		PerPage: req.PerPage,
		Total:   total,
	})
}

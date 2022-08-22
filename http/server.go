package http

import (
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/echovl/orderflo-dev/db"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Config holds the server settings
type Config struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	Core      *layerhub.Core
	SessionDB db.KeyValueDB
}

// Server manages the HTTP implementation of this API
type Server struct {
	App       *fiber.App
	Core      *layerhub.Core
	validate  *validator.Validate
	sessionDB db.KeyValueDB
}

// NewServer creates a new server instance
func NewServer(conf Config) *Server {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name

	})

	srv := &Server{
		Core:      conf.Core,
		sessionDB: conf.SessionDB,
		validate:  validate,
	}

	srv.App = fiber.New(fiber.Config{
		ErrorHandler:          srv.errorHandler,
		ReadTimeout:           conf.ReadTimeout,
		WriteTimeout:          conf.WriteTimeout,
		IdleTimeout:           conf.IdleTimeout,
		DisableStartupMessage: true,
	})

	return srv
}

// ListenAndServe serves HTTP requests from the given addr
func (s *Server) ListenAndServe(addr string) error {
	s.initRoutes()

	errs := make(chan error)
	go func() { errs <- s.App.Listen(addr) }()
	go func() { errs <- s.gracefulShutdown() }()

	return <-errs
}

func (s *Server) gracefulShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return s.App.Shutdown()
}

var (
	MalformedRequestError = errors.E(errors.KindValidation, "request is malformed or invalid")
	MissingSessionError   = errors.E(errors.KindUnexpected, "missing session")
)

func (s *Server) loggerHandler(c *fiber.Ctx) error {
	s.Core.Logger.Infow("",
		"ip", c.IP(),
		"method", c.Method(),
		"path", c.Path(),
	)

	return c.Next()
}

func (s *Server) errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	s.Core.Logger.Error(err.Error())

	var message string
	switch true {
	case errors.Is(err, errors.KindNotFound):
		code = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, errors.KindValidation):
		code = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, errors.KindAuthentication):
		code = http.StatusUnauthorized
		message = "Authentication required"
	case errors.Is(err, errors.KindAuthorization):
		code = http.StatusUnauthorized
		message = "You are not authorized to perform this action"
	default:
		// Unexpected error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = err.Error()
		} else {
			code = http.StatusInternalServerError
			message = "Something unexpected went wrong. Please contact support"
		}
	}

	return c.Status(code).JSON(map[string]string{
		"error": message,
	})
}

// requestParser binds the request body and query to a struct and validates it
func (s *Server) requestParser(c *fiber.Ctx, out any) error {
	err := c.BodyParser(out)
	if err != nil {
		if !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return err
		}
	}
	err = c.QueryParser(out)
	if err != nil {
		return err
	}

	err = s.validate.Struct(out)
	if err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			return errors.Errorf("'%s' with value '%v' failed the '%s' validation", errs[0].Namespace(), errs[0].Value(), errs[0].Tag())
		}
	}

	return nil
}

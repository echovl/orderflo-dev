package http

import (
	"encoding/json"
	"time"

	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
)

const (
	sessionMaxDuration = 30 * 24 * time.Hour
	csrfTokenLength    = 60
)

type Session struct {
	User      *layerhub.User     `json:"user"`
	Company   *layerhub.Company  `json:"company"`
	Customer  *layerhub.Customer `json:"customer"`
	CSRFToken string             `json:"csfr_token"`
}

func (s *Server) getSession(c *fiber.Ctx) (*Session, error) {
	if sess, ok := c.Locals("session").(*Session); ok && sess != nil {
		return sess, nil
	}

	sessID := c.Cookies(sessionCookieName)
	if sessID == "" {
		return nil, errors.Authentication("empty session id")
	}

	sessJSON, err := s.sessionDB.Get(c.Context(), sessID)
	if err != nil {
		return nil, errors.Authentication(errors.Errorf("getting session: %s", err))
	}

	var sess Session
	err = json.Unmarshal(sessJSON, &sess)
	if err != nil {
		return nil, errors.Authentication(errors.Errorf("unmarshaling: %s", err))
	}

	return &sess, nil
}

// initSession creates the user session and sets the session cookie.
// Returns the csrf token to include in the following requests
func (s *Server) initUserSession(c *fiber.Ctx, user *layerhub.User, company *layerhub.Company) (string, error) {
	sessID := layerhub.UniqueID("sess")
	sess := &Session{
		User:      user,
		Company:   company,
		CSRFToken: layerhub.RandomString(csrfTokenLength),
	}

	sessJSON, err := json.Marshal(sess)
	if err != nil {
		return "", err
	}

	err = s.sessionDB.Set(c.Context(), sessID, sessJSON, sessionMaxDuration)
	if err != nil {
		return "", err
	}

	c.Cookie(&fiber.Cookie{
		Name:  sessionCookieName,
		Value: sessID,
	})

	return sess.CSRFToken, nil
}

func (s *Server) initCustomerSession(c *fiber.Ctx, customer *layerhub.Customer, company *layerhub.Company) (string, error) {
	sessID := layerhub.UniqueID("sess")
	sess := &Session{
		Company:   company,
		Customer:  customer,
		CSRFToken: layerhub.RandomString(csrfTokenLength),
	}

	sessJSON, err := json.Marshal(sess)
	if err != nil {
		return "", err
	}

	err = s.sessionDB.Set(c.Context(), sessID, sessJSON, sessionMaxDuration)
	if err != nil {
		return "", err
	}

	c.Cookie(&fiber.Cookie{
		Name:  sessionCookieName,
		Value: sessID,
	})

	return sess.CSRFToken, nil
}

func (s *Server) cleanSession(c *fiber.Ctx) error {
	sessID := c.Cookies(sessionCookieName)
	if sessID == "" {
		return errors.Authentication("nothing to clean, empty session id")
	}

	err := s.sessionDB.Del(c.Context(), sessID)
	if err != nil {
		return errors.Authentication(errors.Errorf("cleaning session: %s", err))
	}

	c.Cookie(&fiber.Cookie{
		Name:  sessionCookieName,
		Value: "",
	})

	return nil
}

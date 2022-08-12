package http

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
)

const (
	sessionMaxDuration = 30 * 24 * time.Hour
	csrfTokenLength    = 60
)

type Session struct {
	UserID    string            `json:"user_id"`
	CSRFToken string            `json:"csfr_token"`
	Role      layerhub.UserRole `json:"role"`
}

func (s *Session) IsAdmin() bool {
	return s.Role == layerhub.UserAdmin
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
func (s *Server) initSession(c *fiber.Ctx, user *layerhub.User) (string, error) {
	sessID := layerhub.UniqueID("sess")

	sess := &Session{
		UserID:    user.ID,
		Role:      user.Role,
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

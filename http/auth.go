package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/layerhub-io/api/assign"
	"github.com/layerhub-io/api/errors"
	"github.com/layerhub-io/api/layerhub"
	"github.com/segmentio/analytics-go"
)

const (
	oauth2StateCookieName  = "auth-state"
	oauth2StateTokenLength = 60
	sessionCookieName      = "auth-session"
	csrfHeaderName         = "Auth-Csrf-Token"
)

func (s *Server) requireUserSession(c *fiber.Ctx) error {
	session, err := s.getSession(c)
	if err != nil {
		return err
	}

	switch c.Method() {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		headers := c.GetReqHeaders()
		csrfToken, ok := headers[csrfHeaderName]
		if !ok {
			return errors.Authentication("missing csrf token")
		}

		if csrfToken != session.CSRFToken {
			return errors.Authentication("mismatched csrf tokens")
		}
	default:
	}

	c.Locals("session", session)

	user, err := s.Core.GetUser(c.Context(), session.UserID)
	if err != nil {
		return errors.Authentication(err)
	}

	s.segment.Enqueue(analytics.Identify{
		UserId: session.UserID,
		Traits: analytics.NewTraits().
			SetFirstName(user.FirstName).
			SetLastName(user.LastName).
			SetEmail(user.Email),
	})

	return c.Next()
}

func (s *Server) requireAppSession(c *fiber.Ctx) error {
	authHeader := c.GetReqHeaders()["Authorization"]
	if authHeader == "" {
		return errors.Authentication("empty authorization header")
	}

	token := strings.ReplaceAll(authHeader, "Bearer ", "")

	users, err := s.Core.FindUsers(context.TODO(), &layerhub.Filter{ApiToken: token, Limit: 1})
	if err != nil {
		return errors.Authentication(err)
	}
	if len(users) == 0 {
		return errors.Authentication("user not found")
	}

	c.Locals("session", &Session{UserID: users[0].ID, Role: users[0].Role})

	return c.Next()
}

func (s *Server) requireAdmin(c *fiber.Ctx) error {
	session, _ := s.getSession(c)
	if !session.IsAdmin() {
		return fiber.NewError(401, "Unauthorized!")
	}

	return c.Next()
}

type User struct {
	*layerhub.User
	PasswordHash string `json:"password_hash,omitempty"`
	PlanID       string `json:"plan_id,omitempty"`
}

func (s *Server) handleGetCSRFToken(c *fiber.Ctx) error {
	type response struct {
		CSRFToken string `json:"csrf_token"`
	}

	sess, _ := s.getSession(c)

	return c.JSON(response{sess.CSRFToken})
}

func (s *Server) handleSignUp(c *fiber.Ctx) error {
	type request struct {
		FirstName string `json:"first_name" validate:"max=20"`
		LastName  string `json:"last_name" validate:"max=20"`
		Email     string `json:"email" validate:"email"`
		Phone     string `json:"phone"`
		Avatar    string `json:"avatar"`
		Company   string `json:"company"`
		Password  string `json:"password" validate:"required,min=8"`
	}

	type response struct {
		User User `json:"user"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	user := layerhub.NewUser()
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.Phone = req.Phone
	user.Avatar = req.Avatar
	user.Company = req.Company

	err := s.Core.RegisterUser(c.Context(), user, req.Password)
	if err != nil {
		return err
	}

	resp := response{
		User: User{User: user},
	}

	s.segment.Enqueue(analytics.Track{
		Event:  "Sign In",
		UserId: user.ID,
	})

	return c.JSON(resp)
}

func (s *Server) handleSignIn(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email" validate:"email"`
		Password string `json:"password" validate:"required"`
	}

	type response struct {
		User      User   `json:"user"`
		CSRFToken string `json:"csrf_token"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	user, err := s.Core.LoginUser(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	csrfToken, err := s.initSession(c, user)
	if err != nil {
		return err
	}

	resp := response{
		User:      User{User: user},
		CSRFToken: csrfToken,
	}

	s.segment.Enqueue(analytics.Track{
		Event:  "Sign Up",
		UserId: user.ID,
	})

	return c.JSON(resp)
}

func (s *Server) handleSignOut(c *fiber.Ctx) error {
	err := s.cleanSession(c)
	if err != nil {
		return err
	}

	return c.SendString("ok")
}

func (s *Server) handleUpdateUserProfile(c *fiber.Ctx) error {
	type request struct {
		FirstName string `json:"first_name" validate:"max=20"`
		LastName  string `json:"last_name" validate:"max=20"`
		Avatar    string `json:"avatar"`
		Company   string `json:"company"`
	}

	type response struct {
		User User `json:"user"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	session, _ := s.getSession(c)
	user, err := s.Core.GetUser(c.Context(), session.UserID)
	if err != nil {
		return err
	}

	if err := assign.Structs(user, req); err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	user.UpdatedAt = time.Now()
	err = s.Core.PutUser(c.Context(), user)
	if err != nil {
		return err
	}

	resp := response{
		User: User{User: user},
	}

	return c.JSON(resp)
}

func (s *Server) handleGoogleSignIn(c *fiber.Ctx) error {
	state := createAuthState()

	googleURL := s.Core.GoogleAuthURL(c.Context(), state)

	c.Cookie(&fiber.Cookie{
		Name:  oauth2StateCookieName,
		Value: state,
	})

	return c.Redirect(googleURL, 302)
}

func (s *Server) handleGithubSignIn(c *fiber.Ctx) error {
	state := createAuthState()

	githubURL := s.Core.GithubAuthURL(c.Context(), state)

	c.Cookie(&fiber.Cookie{
		Name:  oauth2StateCookieName,
		Value: state,
	})

	return c.Redirect(githubURL, 302)
}

func (s *Server) handleGithubCallback(c *fiber.Ctx) error {
	type request struct {
		Code  string `query:"code"`
		State string `query:"state"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	sessionState := c.Cookies(oauth2StateCookieName)

	if req.State != sessionState {
		return errors.Authentication("mismatched auth state")
	}

	user, err := s.Core.LoginGithubUser(c.Context(), req.Code, req.State)
	if err != nil {
		return err
	}

	_, err = s.initSession(c, user)
	if err != nil {
		return err
	}

	cleanOauth2Cookies(c)

	return c.Redirect("https://app.layerhub.io", 302)
}

func (s *Server) handleGoogleCallback(c *fiber.Ctx) error {
	type request struct {
		Code  string `query:"code"`
		State string `query:"state"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	sessionState := c.Cookies(oauth2StateCookieName)

	if req.State != sessionState {
		return errors.Authentication("mismatched auth state")
	}

	user, err := s.Core.LoginGoogleUser(c.Context(), req.Code, req.State)
	if err != nil {
		return err
	}

	_, err = s.initSession(c, user)
	if err != nil {
		return err
	}

	cleanOauth2Cookies(c)

	return c.Redirect("https://app.layerhub.io", 302)
}

func (s *Server) handleCurrentUser(c *fiber.Ctx) error {
	type response struct {
		User User `json:"user"`
	}

	session, _ := s.getSession(c)

	user, err := s.Core.GetUser(c.Context(), session.UserID)
	if err != nil {
		return err
	}

	resp := response{
		User: User{User: user},
	}

	return c.JSON(resp)
}

func cleanOauth2Cookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:  oauth2StateCookieName,
		Value: "",
	})
}

func createAuthState() string {
	return layerhub.RandomString(oauth2StateTokenLength)
}

package http

import (
	"net/http"
	"time"

	"github.com/echovl/orderflo-dev/assign"
	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	"github.com/gofiber/fiber/v2"
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

	if session.User == nil || session.Company == nil {
		return errors.Authentication("mismatched csrf tokens")
	}

	c.Locals("session", session)

	return c.Next()
}

func (s *Server) requireCustomerSession(c *fiber.Ctx) error {
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

	if session.Customer == nil || session.Company == nil {
		return errors.Authentication("mismatched csrf tokens")
	}

	c.Locals("session", session)

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

func (s *Server) handleUserSignUp(c *fiber.Ctx) error {
	type request struct {
		FirstName   string `json:"first_name" validate:"max=20"`
		LastName    string `json:"last_name" validate:"max=20"`
		Email       string `json:"email" validate:"email"`
		Phone       string `json:"phone"`
		Avatar      string `json:"avatar"`
		CompanyName string `json:"company_name"`
		Password    string `json:"password" validate:"required,min=8"`
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
	company := layerhub.NewCompany()
	company.Name = req.CompanyName

	err := s.Core.RegisterUser(c.Context(), user, company, req.Password)
	if err != nil {
		return err
	}

	resp := response{
		User: User{User: user},
	}

	return c.JSON(resp)
}

func (s *Server) handleCustomerSignUp(c *fiber.Ctx) error {
	type request struct {
		FirstName string `json:"first_name" validate:"max=20"`
		LastName  string `json:"last_name" validate:"max=20"`
		Email     string `json:"email" validate:"email"`
		Password  string `json:"password" validate:"required,min=8"`
		CompanyID string `json:"company_id" validate:"required"`
	}

	type response struct {
		Customer *layerhub.Customer `json:"customer"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	customer := layerhub.NewCustomer()
	customer.FirstName = req.FirstName
	customer.LastName = req.LastName
	customer.Email = req.Email
	customer.CompanyID = req.CompanyID

	err := s.Core.RegisterCustomer(c.Context(), customer, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(response{customer})
}

func (s *Server) handleUserSignIn(c *fiber.Ctx) error {
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

	company, err := s.Core.GetCompany(c.Context(), user.CompanyID)
	if err != nil {
		return err
	}

	csrfToken, err := s.initUserSession(c, user, company)
	if err != nil {
		return err
	}

	resp := response{
		User:      User{User: user},
		CSRFToken: csrfToken,
	}

	return c.JSON(resp)
}

func (s *Server) handleCustomerSignIn(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email" validate:"email"`
		Password string `json:"password" validate:"required"`
	}

	type response struct {
		Customer  *layerhub.Customer `json:"customer"`
		CSRFToken string             `json:"csrf_token"`
	}

	var req request
	if err := s.requestParser(c, &req); err != nil {
		return errors.E(errors.KindValidation, err)
	}

	customer, err := s.Core.LoginCustomer(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	company, err := s.Core.GetCompany(c.Context(), customer.CompanyID)
	if err != nil {
		return err
	}

	csrfToken, err := s.initCustomerSession(c, customer, company)
	if err != nil {
		return err
	}

	resp := response{
		Customer:  customer,
		CSRFToken: csrfToken,
	}

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
	user, err := s.Core.GetUser(c.Context(), session.User.ID)
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

	companies, _, err := s.Core.FindCompanies(c.Context(), &layerhub.Filter{UserID: user.ID})
	if err != nil {
		return err
	}

	_, err = s.initUserSession(c, user, &companies[0])
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

	companies, _, err := s.Core.FindCompanies(c.Context(), &layerhub.Filter{UserID: user.ID})
	if err != nil {
		return err
	}

	_, err = s.initUserSession(c, user, &companies[0])
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

	user, err := s.Core.GetUser(c.Context(), session.User.ID)
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

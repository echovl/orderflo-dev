package layerhub

import (
	"context"

	"github.com/echovl/orderflo-dev/errors"
)

type AuthSource string

const (
	AuthSourceEmail  AuthSource = "email"
	AuthSourceGithub AuthSource = "github"
	AuthSourceGoogle AuthSource = "google"
)

func (c *Core) GoogleAuthURL(ctx context.Context, state string) string {
	return c.google.AuthURL(state)
}

func (c *Core) GithubAuthURL(ctx context.Context, state string) string {
	return c.github.AuthURL(state)
}

func (c *Core) RegisterUser(ctx context.Context, user *User, company *Company, password string) error {
	users, err := c.db.FindUsers(ctx, &Filter{Email: user.Email, AuthSource: AuthSourceEmail})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	if len(users) != 0 {
		return errors.E(errors.KindValidation, "email already taken")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	user.PasswordHash = hash
	user.CompanyID = company.ID
	if err := c.db.PutUser(ctx, user); err != nil {
		return err
	}

	if err := c.db.PutCompany(ctx, company); err != nil {
		return err
	}

	return nil
}

func (c *Core) RegisterCustomer(ctx context.Context, customer *Customer, password string) error {
	customers, err := c.db.FindUsers(ctx, &Filter{Email: customer.Email, AuthSource: AuthSourceEmail})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	if len(customers) != 0 {
		return errors.E(errors.KindValidation, "email already taken")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	customer.PasswordHash = hash
	if err := c.db.PutCustomer(ctx, customer); err != nil {
		return err
	}

	return nil
}

func (c *Core) LoginUser(ctx context.Context, email, password string) (*User, error) {
	users, err := c.db.FindUsers(ctx, &Filter{Email: email, AuthSource: AuthSourceEmail})
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	if len(users) == 0 {
		return nil, errors.E(errors.KindValidation, "email not registered")
	}

	if ok := compareHashAndPassword(users[0].PasswordHash, password); !ok {
		return nil, errors.E(errors.KindValidation, "mismatched password")
	}

	return &users[0], nil
}

func (c *Core) LoginCustomer(ctx context.Context, email, password string) (*Customer, error) {
	customers, err := c.db.FindCustomers(ctx, &Filter{Email: email, AuthSource: AuthSourceEmail})
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	if len(customers) == 0 {
		return nil, errors.E(errors.KindValidation, "email not registered")
	}

	if ok := compareHashAndPassword(customers[0].PasswordHash, password); !ok {
		return nil, errors.E(errors.KindValidation, "mismatched password")
	}

	return &customers[0], nil
}

func (c *Core) LoginGithubUser(ctx context.Context, code, state string) (*User, error) {
	client, err := c.github.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	githubUser, err := c.github.CurrentUser(ctx, client)
	if err != nil {
		return nil, err
	}

	users, err := c.db.FindUsers(ctx, &Filter{
		Email:      githubUser.Email,
		AuthSource: AuthSourceGithub,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		user := NewUser()
		user.Email = githubUser.Email
		user.FirstName = githubUser.Name
		user.Avatar = githubUser.AvatarURL
		user.Source = AuthSourceGithub

		err := c.db.PutUser(ctx, user)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return &users[0], nil
}

func (c *Core) LoginGoogleUser(ctx context.Context, code, state string) (*User, error) {
	client, err := c.google.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	googleUser, err := c.google.CurrentUser(ctx, client)
	if err != nil {
		return nil, err
	}

	users, err := c.db.FindUsers(ctx, &Filter{
		Email:      googleUser.Email,
		AuthSource: AuthSourceGoogle,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		user := NewUser()
		user.Email = googleUser.Email
		user.FirstName = googleUser.GivenName
		user.LastName = googleUser.FamilyName
		user.Avatar = googleUser.Picture
		user.Source = AuthSourceGoogle

		err := c.db.PutUser(ctx, user)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return &users[0], nil
}

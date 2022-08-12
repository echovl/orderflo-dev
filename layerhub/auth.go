package layerhub

import (
	"context"

	"github.com/layerhub-io/api/errors"
)

func (c *Core) GoogleAuthURL(ctx context.Context, state string) string {
	return c.google.AuthURL(state)
}

func (c *Core) GithubAuthURL(ctx context.Context, state string) string {
	return c.github.AuthURL(state)
}

func (c *Core) RegisterUser(ctx context.Context, user *User, password string) error {
	users, err := c.db.FindUsers(ctx, &Filter{Email: user.Email, UserSource: UserSourceEmail})
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
	if err := c.db.PutUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (c *Core) LoginUser(ctx context.Context, email, password string) (*User, error) {
	users, err := c.db.FindUsers(ctx, &Filter{Email: email, UserSource: UserSourceEmail})
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
		UserSource: UserSourceGithub,
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
		user.Source = UserSourceGithub

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
		UserSource: UserSourceGoogle,
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
		user.Source = UserSourceGoogle

		err := c.db.PutUser(ctx, user)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	return &users[0], nil
}

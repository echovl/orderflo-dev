package layerhub

import (
	"context"
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	appTokenLength = 60
)

type UserKind string

const (
	UserKindAdmin   UserKind = "admin"
	UserKindCompany UserKind = "company"
)

type UserSource string

const (
	UserSourceEmail  UserSource = "email"
	UserSourceGithub UserSource = "github"
	UserSourceGoogle UserSource = "google"
)

type User struct {
	ID            string     `json:"id" db:"id"`
	FirstName     string     `json:"first_name" db:"first_name"`
	LastName      string     `json:"last_name" db:"last_name"`
	Email         string     `json:"email" db:"email"`
	Phone         string     `json:"phone" db:"phone"`
	Avatar        string     `json:"avatar" db:"avatar"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified bool       `json:"phone_verified" db:"phone_verified"`
	PasswordHash  string     `json:"password_hash" db:"password_hash"`
	PlanID        string     `json:"plan_id" db:"plan_id"`
	Kind          UserKind   `json:"kind" db:"kind"`
	Source        UserSource `json:"source" db:"source"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

func NewUser() *User {
	now := Now()
	return &User{
		ID:        UniqueID("user"),
		Kind:      UserKindCompany,
		Source:    UserSourceEmail,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewApiToken creates a new random token for applications
func NewApiToken() string {
	return RandomString(appTokenLength)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func compareHashAndPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (c *Core) PutUser(ctx context.Context, user *User) error {
	return c.db.PutUser(ctx, user)
}

func (c *Core) GetUser(ctx context.Context, id string) (*User, error) {
	users, err := c.db.FindUsers(ctx, &Filter{ID: id})
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}
	if len(users) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("user '%s' not found", id))
	}

	return &users[0], nil
}

func (c *Core) FindUsers(ctx context.Context, filter *Filter) ([]User, error) {
	return c.db.FindUsers(ctx, filter)
}

package google

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	oauthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenURL = "https://oauth2.googleapis.com/token"
)

type Config struct {
	ClientID    string
	Secret      string
	RedirectURI string
}

type User struct {
	ID         string `json:"sub"`
	FullName   string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Email      string `json:"email"`
}

type Client struct {
	oauth2Config oauth2.Config
}

func NewClient(cfg Config) *Client {
	return &Client{
		oauth2Config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.Secret,
			RedirectURL:  cfg.RedirectURI,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  oauthURL,
				TokenURL: tokenURL,
			},
		},
	}
}

func (c *Client) AuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state)
}

func (c *Client) Exchange(ctx context.Context, code string) (*http.Client, error) {
	token, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return c.oauth2Config.Client(ctx, token), nil
}

func (c *Client) CurrentUser(ctx context.Context, client *http.Client) (*User, error) {
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

package github

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	baseOAuthURL = "https://github.com/login/oauth"
	baseApiURL   = "https://api.github.com"
)

type Config struct {
	ClientID    string
	Secret      string
	RedirectURI string
}

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
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
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				AuthURL:  baseOAuthURL + "/authorize",
				TokenURL: baseOAuthURL + "/access_token",
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
	resp, err := client.Get(baseApiURL + "/user")
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

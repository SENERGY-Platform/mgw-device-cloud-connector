package auth_client

import (
	"context"
	"errors"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const authPath = "auth/realms/master/protocol/openid-connect/token"

type openidToken struct {
	AccessToken      string  `json:"access_token"`
	ExpiresIn        float64 `json:"expires_in"`
	RefreshToken     string  `json:"refresh_token"`
	RefreshExpiresIn float64 `json:"refresh_expires_in"`
	//TokenType        string  `json:"token_type"`
	requestTime time.Time
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
	user       string
	pw         string
	clientID   string
	token      *openidToken
	mu         sync.RWMutex
}

func New(httpClient base_client.HTTPClient, baseUrl, user, password, clientID string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
		user:       user,
		pw:         password,
		clientID:   clientID,
	}
}

func (c *Client) SetHeader(req *http.Request) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token != nil && time.Since(c.token.requestTime).Seconds() < c.token.ExpiresIn {
		req.Header.Set("Bearer", c.token.AccessToken)
		return true
	}
	return false
}

func (c *Client) SetToken(ctx context.Context) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != nil {
		if time.Since(c.token.requestTime).Seconds() >= c.token.ExpiresIn {
			if time.Since(c.token.requestTime).Seconds() >= c.token.RefreshExpiresIn {
				err = c.getToken(ctx)
			} else {
				err = c.refreshToken(ctx)
			}
		}
	} else {
		err = c.getToken(ctx)
	}
	return
}

func (c *Client) getToken(ctx context.Context) (err error) {
	c.token, err = c.tokenRequest(ctx, url.Values{
		"grant_type": {"password"},
		"username":   {c.user},
		"password":   {c.pw},
		"client_id":  {c.clientID},
	})
	return
}

func (c *Client) refreshToken(ctx context.Context) (err error) {
	if c.token == nil {
		return errors.New("missing refresh token")
	}
	c.token, err = c.tokenRequest(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.token.RefreshToken},
		"client_id":     {c.clientID},
	})
	return
}

func (c *Client) tokenRequest(ctx context.Context, form url.Values) (*openidToken, error) {
	u, err := url.JoinPath(c.baseUrl, authPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	token := openidToken{requestTime: time.Now()}
	err = c.baseClient.ExecRequestJSON(req, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func customError(code int, err error) error {
	switch code {
	case http.StatusInternalServerError:
		err = NewInternalError(err)
	case http.StatusNotFound:
		err = NewNotFoundError(err)
	}
	return err
}

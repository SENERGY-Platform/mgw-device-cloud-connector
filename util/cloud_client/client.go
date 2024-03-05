package cloud_client

import (
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/auth_client"
	"net/http"
)

type ClientItf interface {
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
	authClient *auth_client.Client
}

func New(httpClient base_client.HTTPClient, baseUrl string, authClient *auth_client.Client) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
		authClient: authClient,
	}
}

func (c *Client) setAuthHeader(ctx context.Context, req *http.Request) error {
	if c.authClient != nil {
		if !c.authClient.SetHeader(req) {
			err := c.authClient.SetToken(ctx)
			if err != nil {
				return err
			}
			c.authClient.SetHeader(req)
		}
	}
	return nil
}

func customError(code int, err error) error {
	switch code {
	case http.StatusInternalServerError:
		err = NewInternalError(err)
	case http.StatusNotFound:
		err = NewNotFoundError(err)
	case http.StatusUnauthorized:
		err = NewUnauthorizedError(err)
	}
	return err
}

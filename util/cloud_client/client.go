package cloud_client

import (
	"context"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/auth_client"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

type ClientItf interface {
	CreateHub(ctx context.Context, hub models.Hub) (string, error)
	GetHub(ctx context.Context, id string) (models.Hub, error)
	UpdateHub(ctx context.Context, hub models.Hub) error
	CreateDevice(ctx context.Context, device models.Device) (string, error)
	GetDevice(ctx context.Context, id string) (models.Device, error)
	GetDeviceL(ctx context.Context, id string) (models.Device, error)
	UpdateDevice(ctx context.Context, device models.Device, attributeOrigin string) error
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
		err = newInternalError(err)
	case http.StatusNotFound:
		err = newNotFoundError(err)
	case http.StatusUnauthorized:
		err = newUnauthorizedError(err)
	case http.StatusBadRequest:
		err = newBadRequestError(err)
	case http.StatusForbidden:
		err = newForbiddenError(err)
	}
	return err
}

package cloud_client

import (
	"context"
	base_client "github.com/SENERGY-Platform/go-base-http-client"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

type ClientItf interface {
	GetHub(ctx context.Context, id string) (models.Hub, error)
	UpdateHub(ctx context.Context, hub models.Hub) error
	CreateDevice(ctx context.Context, device models.Device) (string, error)
	GetDevices(ctx context.Context, ids []string) ([]models.Device, error)
	GetDevicesL(ctx context.Context, ids []string) ([]models.Device, error)
	UpdateDevice(ctx context.Context, device models.Device, attributeOrigin string) error
	GetAccessPolicies(ctx context.Context) (EndpointAccPolItf, error)
}

type HttpMethodAccPolItf interface {
	Read() bool
	Create() bool
	Update() bool
	Delete() bool
}

type EndpointAccPolItf interface {
	Hubs() HttpMethodAccPolItf
	Devices() HttpMethodAccPolItf
	DevicesL() HttpMethodAccPolItf
}

type Client struct {
	baseClient *base_client.Client
	baseUrl    string
}

func New(httpClient base_client.HTTPClient, baseUrl string) *Client {
	return &Client{
		baseClient: base_client.New(httpClient, customError, ""),
		baseUrl:    baseUrl,
	}
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
	case http.StatusMethodNotAllowed:
		err = newNotAllowedError(err)
	}
	return err
}

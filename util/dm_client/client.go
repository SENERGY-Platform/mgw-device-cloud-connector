package dm_client

import (
	"context"
	"github.com/SENERGY-Platform/go-base-http-client"
	"net/http"
)

type ClientItf interface {
	GetDevices(ctx context.Context) (map[string]Device, error)
	GetDevice(ctx context.Context, id string) (Device, error)
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
	}
	return err
}

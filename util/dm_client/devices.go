package dm_client

import (
	"context"
	"net/http"
	"net/url"
)

const devicesPath = "devices"

func (c *Client) GetDevices(ctx context.Context) (map[string]Device, error) {
	u, err := url.JoinPath(c.baseUrl, devicesPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var devices map[string]Device
	err = c.baseClient.ExecRequestJSON(req, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) GetDevice(ctx context.Context, id string) (Device, error) {
	u, err := url.JoinPath(c.baseUrl, devicesPath, id)
	if err != nil {
		return Device{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Device{}, err
	}
	var device Device
	err = c.baseClient.ExecRequestJSON(req, &device)
	if err != nil {
		return Device{}, err
	}
	return device, nil
}

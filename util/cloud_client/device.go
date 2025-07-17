package cloud_client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
	"strings"
)

const (
	devicesPath      = "device-manager/devices"
	localDevicesPath = "device-manager/local-devices"
)

func (c *Client) CreateDevice(ctx context.Context, device models.Device) (string, error) {
	u, err := url.JoinPath(c.baseUrl, devicesPath)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(device)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	var d models.Device
	err = c.baseClient.ExecRequestJSON(req, &d)
	if err != nil {
		return "", err
	}
	return d.Id, nil
}

func (c *Client) GetDevice(ctx context.Context, id string) (models.Device, error) {
	u, err := url.JoinPath(c.baseUrl, devicesPath, url.QueryEscape(id))
	if err != nil {
		return models.Device{}, err
	}
	return c.getDevice(ctx, u)
}

func (c *Client) GetDevices(ctx context.Context, ids []string) ([]models.Device, error) {
	u, err := url.JoinPath(c.baseUrl, devicesPath)
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		u += "?ids=" + strings.Join(ids, ",")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var devices []models.Device
	err = c.baseClient.ExecRequestJSON(req, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) GetDeviceL(ctx context.Context, id string) (models.Device, error) {
	u, err := url.JoinPath(c.baseUrl, localDevicesPath, url.QueryEscape(id))
	if err != nil {
		return models.Device{}, err
	}
	return c.getDevice(ctx, u)
}

func (c *Client) GetDevicesL(ctx context.Context, ids []string) ([]models.Device, error) {
	u, err := url.JoinPath(c.baseUrl, localDevicesPath)
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		u += "?ids=" + strings.Join(ids, ",")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	var devices []models.Device
	err = c.baseClient.ExecRequestJSON(req, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) UpdateDevice(ctx context.Context, device models.Device, attributeOrigin string) error {
	u, err := url.JoinPath(c.baseUrl, devicesPath, url.QueryEscape(device.Id))
	if err != nil {
		return err
	}
	if attributeOrigin != "" {
		u += "?update-only-same-origin-attributes=" + attributeOrigin
	}
	body, err := json.Marshal(device)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	return c.baseClient.ExecRequestVoid(req)
}

func (c *Client) getDevice(ctx context.Context, u string) (models.Device, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return models.Device{}, err
	}
	var device models.Device
	err = c.baseClient.ExecRequestJSON(req, &device)
	if err != nil {
		return models.Device{}, err
	}
	return device, nil
}

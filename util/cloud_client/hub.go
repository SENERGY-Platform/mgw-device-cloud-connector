package cloud_client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
)

const hubsPath = "device-manager/hubs"

func (c *Client) CreateHub(ctx context.Context, hub models.Hub) (string, error) {
	u, err := url.JoinPath(c.baseUrl, hubsPath)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(hub)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	if err = c.setAuthHeader(ctx, req); err != nil {
		return "", err
	}
	var h models.Hub
	err = c.baseClient.ExecRequestJSON(req, &h)
	if err != nil {
		return "", err
	}
	return h.Id, nil
}

func (c *Client) GetHub(ctx context.Context, id string) (models.Hub, error) {
	u, err := url.JoinPath(c.baseUrl, hubsPath, url.QueryEscape(id))
	if err != nil {
		return models.Hub{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return models.Hub{}, err
	}
	if err = c.setAuthHeader(ctx, req); err != nil {
		return models.Hub{}, err
	}
	var hub models.Hub
	err = c.baseClient.ExecRequestJSON(req, &hub)
	if err != nil {
		return models.Hub{}, err
	}
	return hub, nil
}

func (c *Client) UpdateHub(ctx context.Context, hub models.Hub) error {
	u, err := url.JoinPath(c.baseUrl, hubsPath, url.QueryEscape(hub.Id))
	if err != nil {
		return err
	}
	body, err := json.Marshal(hub)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if err = c.setAuthHeader(ctx, req); err != nil {
		return err
	}
	return c.baseClient.ExecRequestVoid(req)
}

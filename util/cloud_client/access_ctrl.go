package cloud_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const accPolPath = "ladon/allowed"

var httpMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
var httpMethodsLen = len(httpMethods)

type reqItem struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

type resp struct {
	Allowed []bool `json:"allowed"`
}

type HttpMethodAccPol struct {
	get    bool
	head   bool
	post   bool
	put    bool
	patch  bool
	delete bool
}

func (a HttpMethodAccPol) Get() bool {
	return a.get
}

func (a HttpMethodAccPol) Post() bool {
	return a.post
}

func (a HttpMethodAccPol) Put() bool {
	return a.put
}

func (a HttpMethodAccPol) Patch() bool {
	return a.patch
}

func (a HttpMethodAccPol) Delete() bool {
	return a.delete
}

func (c *Client) getAccPol(ctx context.Context, endpoint string) (HttpMethodAccPol, error) {
	u, err := url.JoinPath(c.baseUrl, accPolPath)
	if err != nil {
		return HttpMethodAccPol{}, err
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	var reqData []reqItem
	for _, method := range httpMethods {
		reqData = append(reqData, reqItem{Endpoint: endpoint, Method: method})
	}
	body, err := json.Marshal(reqData)
	if err != nil {
		return HttpMethodAccPol{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return HttpMethodAccPol{}, err
	}
	if err = c.setAuthHeader(ctx, req); err != nil {
		return HttpMethodAccPol{}, err
	}
	var res resp
	err = c.baseClient.ExecRequestJSON(req, &res)
	if err != nil {
		return HttpMethodAccPol{}, err
	}
	fmt.Println(res)
	if httpMethodsLen != len(res.Allowed) {
		return HttpMethodAccPol{}, fmt.Errorf("expected %d results, got %d", httpMethodsLen, len(res.Allowed))
	}
	return HttpMethodAccPol{
		get:    res.Allowed[0],
		head:   res.Allowed[1],
		post:   res.Allowed[2],
		put:    res.Allowed[3],
		patch:  res.Allowed[4],
		delete: res.Allowed[5],
	}, nil
}

func (c *Client) GetDevicesAccPol(ctx context.Context) (HttpMethodAccPol, error) {
	return c.getAccPol(ctx, devicesPath)
}

func (c *Client) GetDevicesLAccPol(ctx context.Context) (HttpMethodAccPol, error) {
	return c.getAccPol(ctx, localDevicesPath)
}

func (c *Client) GetHubAccPol(ctx context.Context) (HttpMethodAccPol, error) {
	return c.getAccPol(ctx, hubsPath)
}

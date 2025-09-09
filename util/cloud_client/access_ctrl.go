package cloud_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const accPolPath = "ladon/allowed"

var endpoints = []string{
	hubsPath,
	devicesPath,
	localDevicesPath,
}
var httpMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
var accPolReqData = func() []accPolReqItem {
	var reqData []accPolReqItem
	for _, endpoint := range endpoints {
		for _, method := range httpMethods {
			reqData = append(reqData, accPolReqItem{Endpoint: "/" + endpoint, Method: method})
		}
	}
	return reqData
}()

type accPolReqItem struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
}

type accPolRespData struct {
	Allowed []bool `json:"allowed"`
}

type httpMethodAccPol struct {
	accPol map[string]bool
}

func (a httpMethodAccPol) Read() bool {
	return a.accPol[http.MethodGet]
}

func (a httpMethodAccPol) Create() bool {
	return a.accPol[http.MethodPost]
}

func (a httpMethodAccPol) Update() bool {
	return a.accPol[http.MethodPut]
}

func (a httpMethodAccPol) Delete() bool {
	return a.accPol[http.MethodDelete]
}

type endpointAccPol struct {
	accPol map[string]httpMethodAccPol
}

func (a endpointAccPol) Hubs() HttpMethodAccPolItf {
	return a.accPol[hubsPath]
}

func (a endpointAccPol) Devices() HttpMethodAccPolItf {
	return a.accPol[devicesPath]
}

func (a endpointAccPol) DevicesL() HttpMethodAccPolItf {
	return a.accPol[localDevicesPath]
}

func (c *Client) GetAccessPolicies(ctx context.Context) (EndpointAccPolItf, error) {
	u, err := url.JoinPath(c.baseUrl, accPolPath)
	if err != nil {
		return endpointAccPol{}, err
	}
	body, err := json.Marshal(accPolReqData)
	if err != nil {
		return endpointAccPol{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(body))
	if err != nil {
		return endpointAccPol{}, err
	}
	var res accPolRespData
	err = c.baseClient.ExecRequestJSON(req, &res)
	if err != nil {
		return endpointAccPol{}, err
	}
	if len(endpoints)*len(httpMethods) != len(res.Allowed) {
		return endpointAccPol{}, fmt.Errorf("expected %d results, got %d", len(endpoints)*len(httpMethods), len(res.Allowed))
	}
	eAccPol := endpointAccPol{accPol: make(map[string]httpMethodAccPol)}
	i := 0
	for _, endpoint := range endpoints {
		mAccPol := httpMethodAccPol{accPol: make(map[string]bool)}
		for _, method := range httpMethods {
			mAccPol.accPol[method] = res.Allowed[i]
			i++
		}
		eAccPol.accPol[endpoint] = mAccPol
	}
	return eAccPol, nil
}

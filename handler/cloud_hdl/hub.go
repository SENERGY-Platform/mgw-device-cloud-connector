package cloud_hdl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	context_hdl "github.com/SENERGY-Platform/go-service-base/context-hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"os"
	"path"
)

const dataFile = "data.json"

type data struct {
	HubID          string            `json:"hub_id"`
	DefaultHubName string            `json:"-"`
	DeviceIDMap    map[string]string `json:"device_id_map"` // localID:ID
}

func (h *Handler) Init(ctx context.Context, hubID, hubName string) error {
	if !path.IsAbs(h.wrkSpacePath) {
		return fmt.Errorf("workspace path must be absolute")
	}
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return err
	}
	d, err := readData(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if hubID != "" {
		d.HubID = hubID
	}
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	var deviceIDs []string
	if d.HubID != "" {
		hub, err := h.cloudClient.GetHub(ctxWt, d.HubID)
		if err != nil {
			var nfe *cloud_client.NotFoundError
			if !errors.As(err, &nfe) {
				return err
			}
			ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
			defer cf2()
			d.HubID, err = h.cloudClient.CreateHub(ctxWt2, models.Hub{Name: hubName})
			if err != nil {
				return err
			}
		}
		deviceIDs = hub.DeviceIds
	} else {
		d.HubID, err = h.cloudClient.CreateHub(ctxWt, models.Hub{Name: hubName})
		if err != nil {
			return err
		}
	}
	d.DeviceIDMap, err = h.getDeviceIDMap(ctx, d.DeviceIDMap, deviceIDs)
	if err != nil {
		return err
	}
	h.data = d
	return writeData(h.wrkSpacePath, h.data)
}

func (h *Handler) getDeviceIDMap(ctx context.Context, oldMap map[string]string, deviceIDs []string) (map[string]string, error) {
	deviceIDMap := make(map[string]string)
	if len(deviceIDs) > 0 {
		ch := context_hdl.New()
		defer ch.CancelAll()
		rDeviceIDMap := make(map[string]string)
		for lID, rID := range oldMap {
			rDeviceIDMap[rID] = lID
		}
		for _, rID := range deviceIDs {
			lID, ok := rDeviceIDMap[rID]
			if !ok {
				device, err := h.cloudClient.GetDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), rID)
				if err != nil {
					return nil, err
				}
				lID = device.LocalId
			}
			deviceIDMap[lID] = rID
		}
	}
	return deviceIDMap, nil
}

func readData(p string) (data, error) {
	f, err := os.Open(path.Join(p, dataFile))
	if err != nil {
		return data{}, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var hi data
	if err = d.Decode(&hi); err != nil {
		return data{}, err
	}
	return hi, nil
}

func writeData(p string, hi data) error {
	f, err := os.OpenFile(path.Join(p, dataFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	return e.Encode(hi)
}

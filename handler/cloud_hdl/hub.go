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

const hubFile = "hub.json"

type hub struct {
	ID          string            `json:"id"`
	DeviceIDMap map[string]string `json:"device_id_map"` // localID:ID
}

func (h *Handler) InitHub(ctx context.Context, id, name string) error {
	if !path.IsAbs(h.wrkSpacePath) {
		return fmt.Errorf("workspace path must be absolute")
	}
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return err
	}
	hb, err := readHub(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if id != "" {
		hb.ID = id
	}
	ctxWt, cf := context.WithTimeout(ctx, h.timeout)
	defer cf()
	var deviceIDs []string
	if hb.ID != "" {
		chb, err := h.cloudClient.GetHub(ctxWt, hb.ID)
		if err != nil {
			var nfe *cloud_client.NotFoundError
			if !errors.As(err, &nfe) {
				return err
			}
			ctxWt2, cf2 := context.WithTimeout(ctx, h.timeout)
			defer cf2()
			hb.ID, err = h.cloudClient.CreateHub(ctxWt2, models.Hub{Name: name})
			if err != nil {
				return err
			}
		}
		deviceIDs = chb.DeviceIds
	} else {
		hb.ID, err = h.cloudClient.CreateHub(ctxWt, models.Hub{Name: name})
		if err != nil {
			return err
		}
	}
	hb.DeviceIDMap, err = h.getDeviceIDMap(ctx, hb.DeviceIDMap, deviceIDs)
	if err != nil {
		return err
	}
	h.hub = hb
	return writeHub(h.wrkSpacePath, h.hub)
}

func (h *Handler) getDeviceIDMap(ctx context.Context, oldMap map[string]string, deviceIDs []string) (map[string]string, error) {
	deviceIDMap := make(map[string]string)
	if len(deviceIDs) > 0 {
		ch := context_hdl.New()
		defer ch.CancelAll()
		rDeviceIDMap := make(map[string]string)
		for ldID, dID := range oldMap {
			rDeviceIDMap[dID] = ldID
		}
		for _, dID := range deviceIDs {
			ldID, ok := rDeviceIDMap[dID]
			if !ok {
				device, err := h.cloudClient.GetDevice(ch.Add(context.WithTimeout(ctx, h.timeout)), dID)
				if err != nil {
					return nil, err
				}
				ldID = device.LocalId
			}
			deviceIDMap[ldID] = dID
		}
	}
	return deviceIDMap, nil
}

func readHub(p string) (hub, error) {
	f, err := os.Open(path.Join(p, hubFile))
	if err != nil {
		return hub{}, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var hi hub
	if err = d.Decode(&hi); err != nil {
		return hub{}, err
	}
	return hi, nil
}

func writeHub(p string, hi hub) error {
	f, err := os.OpenFile(path.Join(p, hubFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	return e.Encode(hi)
}

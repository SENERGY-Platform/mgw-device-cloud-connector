package cloud_hdl

import (
	"context"
	"encoding/json"
	"fmt"
	context_hdl "github.com/SENERGY-Platform/go-service-base/context-hdl"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"path"
	"time"
)

const hubFile = "hub.json"

type hubInfo struct {
	ID          string            `json:"id"`
	DeviceIDMap map[string]string `json:"device_id_map"` // localID:ID
}

type Handler struct {
	cloudClient  cloud_client.ClientItf
	mqttClient   mqtt.Client
	timeout      time.Duration
	hubInfo      hubInfo
	wrkSpacePath string
}

func New(cloudClient cloud_client.ClientItf, mqttClient mqtt.Client, timeout time.Duration, wrkSpacePath string) *Handler {
	return &Handler{
		cloudClient:  cloudClient,
		mqttClient:   mqttClient,
		timeout:      timeout,
		wrkSpacePath: wrkSpacePath,
	}
}

func (h *Handler) InitHub(ctx context.Context, id, name string) error {
	if !path.IsAbs(h.wrkSpacePath) {
		return fmt.Errorf("workspace path must be absolute")
	}
	if err := os.MkdirAll(h.wrkSpacePath, 0770); err != nil {
		return err
	}
	hInfo, err := readHubInfo(h.wrkSpacePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if id != "" {
		hInfo.ID = id
	}
	if hInfo.ID != "" {
		hInfo.DeviceIDMap, err = h.refreshDeviceIDMap(ctx, hInfo.ID, hInfo.DeviceIDMap)
		if err != nil {
			return err
		}
	} else {
		ctxWt, cf := context.WithTimeout(ctx, h.timeout)
		defer cf()
		id, err = h.cloudClient.CreateHub(ctxWt, models.Hub{Name: name})
		if err != nil {
			return err
		}
		hInfo.ID = id
	}
	h.hubInfo = hInfo
	return writeHubInfo(h.wrkSpacePath, h.hubInfo)
}

func (h *Handler) refreshDeviceIDMap(ctx context.Context, hID string, oldMap map[string]string) (map[string]string, error) {
	ch := context_hdl.New()
	defer ch.CancelAll()
	hub, err := h.cloudClient.GetHub(ch.Add(context.WithTimeout(ctx, h.timeout)), hID)
	if err != nil {
		return nil, err
	}
	deviceIDMap := make(map[string]string)
	rDeviceIDMap := make(map[string]string)
	for ldID, dID := range oldMap {
		rDeviceIDMap[dID] = ldID
	}
	for _, dID := range hub.DeviceIds {
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
	return deviceIDMap, nil
}

func readHubInfo(p string) (hubInfo, error) {
	f, err := os.Open(path.Join(p, hubFile))
	if err != nil {
		return hubInfo{}, err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var hi hubInfo
	if err = d.Decode(&hi); err != nil {
		return hubInfo{}, err
	}
	return hi, nil
}

func writeHubInfo(p string, hi hubInfo) error {
	f, err := os.OpenFile(path.Join(p, hubFile), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	return e.Encode(hi)
}

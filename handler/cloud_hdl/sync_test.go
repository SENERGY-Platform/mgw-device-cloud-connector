package cloud_hdl

import (
	"context"
	"errors"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
)

func TestHandler_syncDevice(t *testing.T) {
	var mockCC *cloud_client.Mock
	var handler *Handler
	initHandler := func() {
		mockCC = &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
		handler = &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
	}
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	cID := "1"
	lID := "123"
	cDevice := models.Device{
		Id:      cID,
		LocalId: lID,
		Name:    "Test Device",
		Attributes: []models.Attribute{
			{
				Key:    "test-key",
				Value:  "test-value",
				Origin: "test-origin",
			},
		},
		DeviceTypeId: "456",
	}
	lDevice := model.Device{
		ID:    lID,
		Name:  "Test Device",
		State: "online",
		Type:  "456",
		Attributes: []model.Attribute{
			{
				Key:   "test-key",
				Value: "test-value",
			},
		},
	}
	t.Run("cloud device does not exist", func(t *testing.T) {
		initHandler()
		id, err := handler.syncDevice(context.Background(), map[string]models.Device{}, lDevice)
		if err != nil {
			t.Error(err)
		}
		cd, ok := mockCC.Devices[id]
		if !ok {
			t.Error("cloud device ID not in map")
		}
		if !reflect.DeepEqual(cDevice, cd) {
			t.Error("cloud device not equal")
		}
	})
	t.Run("cloud device exists", func(t *testing.T) {
		t.Run("in network equal", func(t *testing.T) {
			initHandler()
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			id, err := handler.syncDevice(context.Background(), map[string]models.Device{
				lID: cDevice,
			}, lDevice)
			if err != nil {
				t.Error(err)
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("illegal call")
			}
			if mockCC.GetDeviceLC > 0 {
				t.Error("illegal call")
			}
			if mockCC.UpdateDeviceC > 0 {
				t.Error("illegal call")
			}
		})
		t.Run("not in network equal", func(t *testing.T) {
			initHandler()
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			id, err := handler.syncDevice(context.Background(), map[string]models.Device{}, lDevice)
			if err != nil {
				t.Error(err)
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.CreateDeviceC != 1 {
				t.Error("missing call")
			}
			if mockCC.GetDeviceLC != 1 {
				t.Error("missing call")
			}
			if mockCC.UpdateDeviceC > 0 {
				t.Error("illegal call")
			}
		})
		lDevice.Name = "test"
		t.Run("in network not equal", func(t *testing.T) {
			initHandler()
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			id, err := handler.syncDevice(context.Background(), map[string]models.Device{
				lID: cDevice,
			}, lDevice)
			if err != nil {
				t.Error(err)
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("illegal call")
			}
			if mockCC.GetDeviceLC > 0 {
				t.Error("illegal call")
			}
			cd := mockCC.Devices[cID]
			if cd.Name != lDevice.Name {
				t.Error("name not equal")
			}
		})
		t.Run("not in network not equal", func(t *testing.T) {
			initHandler()
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			id, err := handler.syncDevice(context.Background(), map[string]models.Device{}, lDevice)
			if err != nil {
				t.Error(err)
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.CreateDeviceC != 1 {
				t.Error("missing call")
			}
			if mockCC.GetDeviceLC != 1 {
				t.Error("missing call")
			}
			cd := mockCC.Devices[cID]
			if cd.Name != lDevice.Name {
				t.Error("name not equal")
			}
		})
	})
	t.Run("request error", func(t *testing.T) {
		initHandler()
		mockCC.Err = errors.New("test error")
		_, err := handler.syncDevice(context.Background(), map[string]models.Device{}, lDevice)
		if err == nil {
			t.Error("error expected")
		}
	})
}

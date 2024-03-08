package cloud_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
)

func TestNewDevice(t *testing.T) {
	a := models.Device{
		Id:      "rid",
		LocalId: "lid",
		Name:    "Test Device",
		Attributes: []models.Attribute{
			{
				Key:    "test-key",
				Value:  "test-val",
				Origin: "test-origin",
			},
		},
		DeviceTypeId: "test-type",
	}
	b := newDevice(model.Device{
		ID:   "lid",
		Name: "Test Device",
		Type: "test-type",
		Attributes: []model.Attribute{
			{
				Key:   "test-key",
				Value: "test-val",
			},
		},
	}, "rid", "test-origin")
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestCreateOrUpdateDevice(t *testing.T) {
	mockCC := &cloud_client.Mock{}
	h := New(mockCC, nil, 0, "", "test-origin")
	t.Run("create device", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = make(map[string]models.Device)
		mockCC.DeviceIDMap = make(map[string]string)
		id, err := h.createOrUpdateDevice(context.Background(), model.Device{
			ID: "test",
		})
		if err != nil {
			t.Error(err)
		}
		if _, ok := mockCC.Devices[id]; !ok {
			t.Error("device not created")
		}
		if mockCC.GetDeviceLC+mockCC.UpdateDeviceC > 0 {
			t.Error("illegal call number")
		}
	})
	t.Run("create device but local id exists", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = map[string]models.Device{
			"1": {
				Id:      "1",
				LocalId: "test",
			},
		}
		mockCC.DeviceIDMap = map[string]string{
			"test": "1",
		}
		id, err := h.createOrUpdateDevice(context.Background(), model.Device{
			ID:   "test",
			Name: "Test Device",
		})
		if err != nil {
			t.Error(err)
		}
		d, ok := mockCC.Devices[id]
		if !ok {
			t.Error("device not created")
		}
		if d.Name != "Test Device" {
			t.Error("device not updated")
		}
		if mockCC.AttributeOrigin != "test-origin" {
			t.Error("invalid attribute origin")
		}
	})
	t.Run("error case", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.SetInternalErr()
		_, err := h.createOrUpdateDevice(context.Background(), model.Device{})
		if err == nil {
			t.Error("error ignored")
		}
	})
}

func TestUpdateOrCreateDevice(t *testing.T) {
	mockCC := &cloud_client.Mock{}
	h := New(mockCC, nil, 0, "", "test-origin")
	t.Run("update device", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = map[string]models.Device{
			"1": {
				Id:      "1",
				LocalId: "test",
			},
		}
		mockCC.DeviceIDMap = map[string]string{
			"test": "1",
		}
		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
			ID:   "test",
			Name: "Test Device",
		})
		if err != nil {
			t.Error(err)
		}
		if id != "1" {
			t.Error("id mismatch")
		}
		d := mockCC.Devices["1"]
		if d.Name != "Test Device" {
			t.Error("device not updated")
		}
		if mockCC.AttributeOrigin != "test-origin" {
			t.Error("invalid attribute origin")
		}
	})
	t.Run("update device but device doesn't exist", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = make(map[string]models.Device)
		mockCC.DeviceIDMap = make(map[string]string)
		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
			ID:   "test",
			Name: "Test Device",
		})
		if err != nil {
			t.Error(err)
		}
		_, ok := mockCC.Devices[id]
		if !ok {
			t.Error("device not created")
		}
	})
	t.Run("update device but local id exists with different remote id", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = map[string]models.Device{
			"2": {
				Id:      "2",
				LocalId: "test",
			},
		}
		mockCC.DeviceIDMap = map[string]string{
			"test": "2",
		}
		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
			ID:   "test",
			Name: "Test Device",
		})
		if err != nil {
			t.Error(err)
		}
		if id != "2" {
			t.Error("id mismatch")
		}
		d := mockCC.Devices["2"]
		if d.Name != "Test Device" {
			t.Error("device not updated")
		}
	})
	t.Run("error case", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.SetInternalErr()
		_, err := h.updateOrCreateDevice(context.Background(), "", model.Device{})
		if err == nil {
			t.Error("error ignored")
		}
	})
}

func TestGetDeviceIDMap(t *testing.T) {
	mockCC := &cloud_client.Mock{}
	h := New(mockCC, nil, 0, "", "")
	t.Run("nil case", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, nil)
		if err != nil {
			return
		}
		if len(deviceIDMap) > 0 {
			t.Error("map not empty")
		}
	})
	t.Run("no device IDs", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		oldMap := map[string]string{
			"test": "1",
		}
		deviceIDMap, err := h.getDeviceIDMap(context.Background(), oldMap, nil)
		if err != nil {
			return
		}
		if len(deviceIDMap) > 0 {
			t.Error("map not empty")
		}
	})
	t.Run("map and device IDs match", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		oldMap := map[string]string{
			"test": "1",
		}
		deviceIDMap, err := h.getDeviceIDMap(context.Background(), oldMap, []string{"1"})
		if err != nil {
			return
		}
		id, ok := deviceIDMap["test"]
		if !ok {
			t.Error("not in map")
		}
		if id != "1" {
			t.Error("id mismatch")
		}
		if mockCC.GetDeviceC > 0 {
			t.Error("illegal call number")
		}
	})
	t.Run("map and device IDs don't match", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = map[string]models.Device{
			"1": {
				Id:      "1",
				LocalId: "test",
			},
		}
		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
		if err != nil {
			return
		}
		id, ok := deviceIDMap["test"]
		if !ok {
			t.Error("not in map")
		}
		if id != "1" {
			t.Error("id mismatch")
		}
		if mockCC.GetDeviceC != 1 {
			t.Error("illegal call number")
		}
	})
	t.Run("map and device IDs don't match and device doesn't exist", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = make(map[string]models.Device)
		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
		if err != nil {
			return
		}
		if len(deviceIDMap) > 0 {
			t.Error("map not empty")
		}
		if mockCC.GetDeviceC != 1 {
			t.Error("illegal call number")
		}
	})
	t.Run("error case", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		mockCC.Devices = map[string]models.Device{
			"1": {
				Id:      "1",
				LocalId: "test",
			},
		}
		mockCC.SetInternalErr()
		_, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
		if err == nil {
			t.Error("error ignored")
		}
	})
}

func TestSyncDevice(t *testing.T) {
	mockCC := &cloud_client.Mock{}
	h := New(mockCC, nil, 0, "", "test-origin")
	hReset := func() {
		h.data.DeviceIDMap = make(map[string]string)
	}
	t.Run("device doesn't exist", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		t.Cleanup(hReset)
		mockCC.Devices = make(map[string]models.Device)
		mockCC.DeviceIDMap = make(map[string]string)
		h.data.DeviceIDMap = make(map[string]string)
		err := h.syncDevice(context.Background(), model.Device{
			ID: "test",
		})
		if err != nil {
			t.Error(err)
		}
		id, ok := h.data.DeviceIDMap["test"]
		if !ok {
			t.Error("local id not in map")
		}
		if _, ok = mockCC.Devices[id]; !ok {
			t.Error("id not in map")
		}
	})
	t.Run("device exists", func(t *testing.T) {
		t.Cleanup(mockCC.Reset)
		t.Cleanup(hReset)
		mockCC.Devices = map[string]models.Device{
			"1": {
				Id:      "1",
				LocalId: "test",
			},
		}
		mockCC.DeviceIDMap = map[string]string{
			"test": "1",
		}
		h.data.DeviceIDMap = map[string]string{
			"test": "1",
		}
		err := h.syncDevice(context.Background(), model.Device{
			ID:   "test",
			Name: "Test Device",
		})
		if err != nil {
			t.Error(err)
		}
		d := mockCC.Devices["1"]
		if d.Name != "Test Device" {
			t.Error("device not updated")
		}
		if mockCC.AttributeOrigin != "test-origin" {
			t.Error("invalid attribute origin")
		}
	})
}

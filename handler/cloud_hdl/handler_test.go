package cloud_hdl

import (
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
)

func Test_newDevice(t *testing.T) {
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

//func TestHandler_createOrUpdateDevice(t *testing.T) {
//	mockCC := &cloud_client.Mock{}
//	h := New(mockCC, 0, "", "test-origin")
//	t.Run("create device", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = make(map[string]models.Device)
//		mockCC.DeviceIDMap = make(map[string]string)
//		id, err := h.createOrUpdateDevice(context.Background(), model.Device{
//			ID: "test",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		if _, ok := mockCC.Devices[id]; !ok {
//			t.Error("device not created")
//		}
//		if mockCC.GetDeviceLC+mockCC.UpdateDeviceC > 0 {
//			t.Error("illegal call number")
//		}
//	})
//	t.Run("create device but local id exists", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "test",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{
//			"test": "1",
//		}
//		id, err := h.createOrUpdateDevice(context.Background(), model.Device{
//			ID:   "test",
//			Name: "Test Device",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		d, ok := mockCC.Devices[id]
//		if !ok {
//			t.Error("device not created")
//		}
//		if d.Name != "Test Device" {
//			t.Error("device not updated")
//		}
//		if mockCC.AttributeOrigin != "test-origin" {
//			t.Error("invalid attribute origin")
//		}
//	})
//	t.Run("error case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.SetInternalErr()
//		_, err := h.createOrUpdateDevice(context.Background(), model.Device{})
//		if err == nil {
//			t.Error("error ignored")
//		}
//	})
//}
//
//func TestHandler_updateOrCreateDevice(t *testing.T) {
//	mockCC := &cloud_client.Mock{}
//	h := New(mockCC, 0, "", "test-origin")
//	t.Run("update device", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "test",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{
//			"test": "1",
//		}
//		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
//			ID:   "test",
//			Name: "Test Device",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		if id != "1" {
//			t.Error("id mismatch")
//		}
//		d := mockCC.Devices["1"]
//		if d.Name != "Test Device" {
//			t.Error("device not updated")
//		}
//		if mockCC.AttributeOrigin != "test-origin" {
//			t.Error("invalid attribute origin")
//		}
//	})
//	t.Run("update device but device doesn't exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = make(map[string]models.Device)
//		mockCC.DeviceIDMap = make(map[string]string)
//		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
//			ID:   "test",
//			Name: "Test Device",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		_, ok := mockCC.Devices[id]
//		if !ok {
//			t.Error("device not created")
//		}
//	})
//	t.Run("update device but local id exists with different remote id", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = map[string]models.Device{
//			"2": {
//				Id:      "2",
//				LocalId: "test",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{
//			"test": "2",
//		}
//		id, err := h.updateOrCreateDevice(context.Background(), "1", model.Device{
//			ID:   "test",
//			Name: "Test Device",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		if id != "2" {
//			t.Error("id mismatch")
//		}
//		d := mockCC.Devices["2"]
//		if d.Name != "Test Device" {
//			t.Error("device not updated")
//		}
//	})
//	t.Run("error case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.SetInternalErr()
//		_, err := h.updateOrCreateDevice(context.Background(), "", model.Device{})
//		if err == nil {
//			t.Error("error ignored")
//		}
//	})
//}
//
//func TestHandler_getDeviceIDMap(t *testing.T) {
//	mockCC := &cloud_client.Mock{}
//	h := New(mockCC, 0, "", "")
//	t.Run("nil case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, nil)
//		if err != nil {
//			return
//		}
//		if len(deviceIDMap) > 0 {
//			t.Error("map not empty")
//		}
//	})
//	t.Run("no device IDs", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		oldMap := map[string]string{
//			"test": "1",
//		}
//		deviceIDMap, err := h.getDeviceIDMap(context.Background(), oldMap, nil)
//		if err != nil {
//			return
//		}
//		if len(deviceIDMap) > 0 {
//			t.Error("map not empty")
//		}
//	})
//	t.Run("map and device IDs match", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		oldMap := map[string]string{
//			"test": "1",
//		}
//		deviceIDMap, err := h.getDeviceIDMap(context.Background(), oldMap, []string{"1"})
//		if err != nil {
//			return
//		}
//		id, ok := deviceIDMap["test"]
//		if !ok {
//			t.Error("not in map")
//		}
//		if id != "1" {
//			t.Error("id mismatch")
//		}
//		if mockCC.GetDeviceC > 0 {
//			t.Error("illegal call number")
//		}
//	})
//	t.Run("map and device IDs don't match", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "test",
//			},
//		}
//		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
//		if err != nil {
//			return
//		}
//		id, ok := deviceIDMap["test"]
//		if !ok {
//			t.Error("not in map")
//		}
//		if id != "1" {
//			t.Error("id mismatch")
//		}
//		if mockCC.GetDeviceC != 1 {
//			t.Error("illegal call number")
//		}
//	})
//	t.Run("map and device IDs don't match and device doesn't exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = make(map[string]models.Device)
//		deviceIDMap, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
//		if err != nil {
//			return
//		}
//		if len(deviceIDMap) > 0 {
//			t.Error("map not empty")
//		}
//		if mockCC.GetDeviceC != 1 {
//			t.Error("illegal call number")
//		}
//	})
//	t.Run("error case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "test",
//			},
//		}
//		mockCC.SetInternalErr()
//		_, err := h.getDeviceIDMap(context.Background(), nil, []string{"1"})
//		if err == nil {
//			t.Error("error ignored")
//		}
//	})
//}
//
//func TestHandler_syncDevice(t *testing.T) {
//	mockCC := &cloud_client.Mock{}
//	h := New(mockCC, 0, "", "test-origin")
//	hReset := func() {
//		h.data.DeviceIDMap = make(map[string]string)
//	}
//	t.Run("device doesn't exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = make(map[string]models.Device)
//		mockCC.DeviceIDMap = make(map[string]string)
//		h.data.DeviceIDMap = make(map[string]string)
//		err := h.syncDevice(context.Background(), model.Device{
//			ID: "test",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		id, ok := h.data.DeviceIDMap["test"]
//		if !ok {
//			t.Error("local id not in map")
//		}
//		if _, ok = mockCC.Devices[id]; !ok {
//			t.Error("id not in map")
//		}
//	})
//	t.Run("device exists", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "test",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{
//			"test": "1",
//		}
//		h.data.DeviceIDMap = map[string]string{
//			"test": "1",
//		}
//		err := h.syncDevice(context.Background(), model.Device{
//			ID:   "test",
//			Name: "Test Device",
//		})
//		if err != nil {
//			t.Error(err)
//		}
//		d := mockCC.Devices["1"]
//		if d.Name != "Test Device" {
//			t.Error("device not updated")
//		}
//		if mockCC.AttributeOrigin != "test-origin" {
//			t.Error("invalid attribute origin")
//		}
//	})
//}
//
//func TestHandler_Sync(t *testing.T) {
//	mockCC := &cloud_client.Mock{}
//	h := New(mockCC, 0, t.TempDir(), "test-origin")
//	hReset := func() {
//		h.data = data{}
//		h.attrOrigin = ""
//	}
//	devices := map[string]model.Device{
//		"a": {
//			ID:   "a",
//			Name: "Test Device A",
//		},
//		"b": {
//			ID:   "b",
//			Name: "Test Device B",
//		},
//	}
//	t.Run("hub and devices don't exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = make(map[string]models.Device)
//		mockCC.DeviceIDMap = make(map[string]string)
//		mockCC.Hubs = make(map[string]models.Hub)
//		h.data = data{
//			DefaultHubName: "Test Hub",
//			DeviceIDMap:    make(map[string]string),
//		}
//		failed, err := h.Sync(context.Background(), devices, nil)
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) > 0 {
//			t.Error("illegal number of failed devices")
//		}
//		for lID, rID := range h.data.DeviceIDMap {
//			d, ok := mockCC.Devices[rID]
//			if !ok {
//				t.Errorf("device '%s' not created", lID)
//			}
//			if d.LocalId != lID {
//				t.Error("id mismatch")
//			}
//		}
//		hub, ok := mockCC.Hubs[h.data.HubID]
//		if !ok {
//			t.Error("hub not created")
//		}
//		if !inSlice("a", hub.DeviceLocalIds) {
//			t.Error("local id 'a' not in map")
//		}
//		if !inSlice("b", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if hub.Name != h.data.DefaultHubName {
//			t.Error("hub default name not set")
//		}
//	})
//	t.Run("hub and devices exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "a",
//			},
//			"2": {
//				Id:      "2",
//				LocalId: "b",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{"a": "1", "b": "2"}
//		mockCC.Hubs = map[string]models.Hub{
//			"test": {
//				Id:             "test",
//				Name:           "Test Hub",
//				Hash:           "",
//				DeviceLocalIds: []string{"a", "b"},
//				DeviceIds:      []string{"1", "2"},
//			},
//		}
//		h.data = data{
//			HubID:       "test",
//			DeviceIDMap: make(map[string]string),
//		}
//		failed, err := h.Sync(context.Background(), devices, nil)
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) > 0 {
//			t.Error("illegal number of failed devices")
//		}
//		rID, ok := h.data.DeviceIDMap["a"]
//		if !ok {
//			t.Error("local id 'a' not in map")
//		}
//		if rID != "1" {
//			t.Error("id mismatch")
//		}
//		rID, ok = h.data.DeviceIDMap["b"]
//		if !ok {
//			t.Error("local id 'b' not in map")
//		}
//		if rID != "2" {
//			t.Error("id mismatch")
//		}
//		hub := mockCC.Hubs[h.data.HubID]
//		if !inSlice("a", hub.DeviceLocalIds) {
//			t.Error("local id 'a' not in map")
//		}
//		if !inSlice("b", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if hub.Name == h.data.DefaultHubName {
//			t.Error("hub default name was set")
//		}
//		if mockCC.CreateDeviceC+mockCC.UpdateDeviceC+mockCC.CreateHubC > 0 {
//			t.Error("illegal number of calls")
//		}
//	})
//	t.Run("only devices exist", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "a",
//			},
//			"2": {
//				Id:      "2",
//				LocalId: "b",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{"a": "1", "b": "2"}
//		mockCC.Hubs = make(map[string]models.Hub)
//		h.data.DeviceIDMap = make(map[string]string)
//		failed, err := h.Sync(context.Background(), devices, nil)
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) > 0 {
//			t.Error("illegal number of failed devices")
//		}
//		hub, ok := mockCC.Hubs[h.data.HubID]
//		if !ok {
//			t.Error("hub not created")
//		}
//		if !inSlice("a", hub.DeviceLocalIds) {
//			t.Error("local id 'a' not in map")
//		}
//		if !inSlice("b", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if hub.Name != h.data.DefaultHubName {
//			t.Error("hub default name not set")
//		}
//	})
//	t.Run("only hub exists", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = make(map[string]models.Device)
//		mockCC.DeviceIDMap = make(map[string]string)
//		mockCC.Hubs = map[string]models.Hub{
//			"test": {
//				Id:   "test",
//				Name: "Test Hub",
//			},
//		}
//		h.data = data{
//			HubID:       "test",
//			DeviceIDMap: make(map[string]string),
//		}
//		failed, err := h.Sync(context.Background(), devices, nil)
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) > 0 {
//			t.Error("illegal number of failed devices")
//		}
//		for lID, rID := range h.data.DeviceIDMap {
//			d, ok := mockCC.Devices[rID]
//			if !ok {
//				t.Errorf("device '%s' not created", lID)
//			}
//			if d.LocalId != lID {
//				t.Error("id mismatch")
//			}
//		}
//		hub := mockCC.Hubs[h.data.HubID]
//		if !inSlice("a", hub.DeviceLocalIds) {
//			t.Error("local id 'a' not in map")
//		}
//		if !inSlice("b", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if hub.Name == h.data.DefaultHubName {
//			t.Error("hub default name was set")
//		}
//	})
//	t.Run("hub exists, devices partially exist and must be updated", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.Devices = map[string]models.Device{
//			"1": {
//				Id:      "1",
//				LocalId: "a",
//				Name:    "Test",
//			},
//			"0": {
//				Id:      "0",
//				LocalId: "c",
//			},
//		}
//		mockCC.DeviceIDMap = map[string]string{"a": "1", "c": "0"}
//		mockCC.Hubs = map[string]models.Hub{
//			"test": {
//				Id:             "test",
//				Name:           "Test Hub",
//				Hash:           "",
//				DeviceLocalIds: []string{"a", "c"},
//				DeviceIds:      []string{"1", "0"},
//			},
//		}
//		h.data = data{
//			HubID:       "test",
//			DeviceIDMap: make(map[string]string),
//		}
//		failed, err := h.Sync(context.Background(), devices, []string{"a"})
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) > 0 {
//			t.Error("illegal number of failed devices")
//		}
//		rID, ok := h.data.DeviceIDMap["a"]
//		if !ok {
//			t.Error("local id 'a' not in map")
//		}
//		if rID != "1" {
//			t.Error("id mismatch")
//		}
//		rID, ok = h.data.DeviceIDMap["b"]
//		if !ok {
//			t.Error("local id 'b' not in map")
//		}
//		rID, ok = h.data.DeviceIDMap["c"]
//		if !ok {
//			t.Error("local id 'c' not in map")
//		}
//		if rID != "0" {
//			t.Error("id mismatch")
//		}
//		d := mockCC.Devices["1"]
//		if d.Name != "Test Device A" {
//			t.Error("device not updated")
//		}
//		hub := mockCC.Hubs[h.data.HubID]
//		if !inSlice("a", hub.DeviceLocalIds) {
//			t.Error("local id 'a' not in map")
//		}
//		if !inSlice("b", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if !inSlice("c", hub.DeviceLocalIds) {
//			t.Error("local id 'b' not in map")
//		}
//		if hub.Name == h.data.DefaultHubName {
//			t.Error("hub default name was set")
//		}
//		if mockCC.CreateHubC > 0 {
//			t.Error("illegal number of calls")
//		}
//	})
//	t.Run("failed devices error case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		util.InitLogger(sb_util.LoggerConfig{Terminal: true})
//		mockCC.SetNotFoundErr()
//		h.data.HubID = "test"
//		failed, err := h.Sync(context.Background(), devices, nil)
//		if err != nil {
//			t.Error(err)
//		}
//		if len(failed) != 2 {
//			t.Error("illegal number of failed devices")
//		}
//	})
//	t.Run("error case", func(t *testing.T) {
//		t.Cleanup(mockCC.Reset)
//		t.Cleanup(hReset)
//		mockCC.SetInternalErr()
//		h.data.HubID = "test"
//		_, err := h.Sync(context.Background(), devices, nil)
//		if err == nil {
//			t.Error(err)
//		}
//	})
//}
//
//func inSlice(s string, sl []string) bool {
//	for _, s2 := range sl {
//		if s2 == s {
//			return true
//		}
//	}
//	return false
//}

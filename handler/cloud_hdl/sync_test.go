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

func TestHandler_syncDevAndNet(t *testing.T) {
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
	hub := models.Hub{
		Id:        "1",
		DeviceIds: []string{cID},
	}
	mockCC := &cloud_client.Mock{
		Devices:     make(map[string]models.Device),
		DeviceIDMap: make(map[string]string),
		Hubs: map[string]models.Hub{
			"1": hub,
		},
		EptAccPol: cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesLAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
		},
	}
	mockCC.Devices[cID] = cDevice
	mockCC.DeviceIDMap[lID] = cID
	handler := &Handler{
		cloudClient: mockCC,
		data:        data{NetworkID: "1"},
		attrOrigin:  "test-origin",
	}
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	ok, err := handler.syncDevAndNet(context.Background(), map[string]model.Device{lID: lDevice})
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("should be true")
	}
	if handler.lastSync.IsZero() {
		t.Error("should not be zero")
	}
	if len(handler.syncedIDs) != 1 {
		t.Error("invalid length")
	}
	t.Run("no device write access rights", func(t *testing.T) {
		mockCC.EptAccPol.DevicesAccPol = cloud_client.HttpMethodAccPolMock{
			ReadAP:   true,
			CreateAP: false,
			UpdateAP: false,
			DeleteAP: false,
		}
		mockCC.EptAccPol.DevicesLAccPol = cloud_client.HttpMethodAccPolMock{
			ReadAP:   true,
			CreateAP: false,
			UpdateAP: false,
			DeleteAP: false,
		}
		ok, err = handler.syncDevAndNet(context.Background(), map[string]model.Device{lID: lDevice})
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Error("should be true")
		}
		if handler.lastSync.IsZero() {
			t.Error("should not be zero")
		}
		if len(handler.syncedIDs) != 1 {
			t.Error("invalid length")
		}
	})
}

func TestHandler_syncNetwork(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("no sync required", func(t *testing.T) {
		hub := models.Hub{
			Id:        "1",
			DeviceIds: []string{"a", "b"},
		}
		mockCC := &cloud_client.Mock{
			Hubs: map[string]models.Hub{
				"1": hub,
			},
		}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		t.Run("number of local devices match devices in network", func(t *testing.T) {
			err := handler.syncNetwork(context.Background(), hub, map[string]string{"1": "a", "2": "b"})
			if err != nil {
				t.Error(err)
			}
			if mockCC.UpdateHubC > 0 {
				t.Error("illegal call")
			}
		})
		t.Run("local devices are a subset of devices in network", func(t *testing.T) {
			err := handler.syncNetwork(context.Background(), hub, map[string]string{"1": "a"})
			if err != nil {
				t.Error(err)
			}
			if mockCC.UpdateHubC > 0 {
				t.Error("illegal call")
			}
		})
	})
	t.Run("sync required", func(t *testing.T) {
		hub := models.Hub{
			Id:        "1",
			DeviceIds: []string{"a", "c"},
		}
		mockCC := &cloud_client.Mock{
			Hubs: map[string]models.Hub{
				"1": hub,
			},
		}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		err := handler.syncNetwork(context.Background(), hub, map[string]string{"1": "a", "2": "b"})
		if err != nil {
			t.Error(err)
		}
		if mockCC.UpdateHubC != 1 {
			t.Error("missing call")
		}
		dIDs := map[string]struct{}{
			"a": {},
			"b": {},
			"c": {},
		}
		h := mockCC.Hubs["1"]
		if len(h.DeviceIds) != 3 {
			t.Error("invalid length")
		}
		for _, id := range h.DeviceIds {
			if _, ok := dIDs[id]; !ok {
				t.Errorf("missing id %s", id)
			}
		}
	})
	t.Run("network not found", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		err := handler.syncNetwork(context.Background(), models.Hub{}, map[string]string{"": ""})
		if err == nil {
			t.Error("error expected")
		}
		if !handler.noNetwork {
			t.Error("true expected")
		}
	})
	t.Run("request error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		mockCC.Err = errors.New("test error")
		err := handler.syncNetwork(context.Background(), models.Hub{}, map[string]string{"": ""})
		if err == nil {
			t.Error("error expected")
		}
		if handler.noNetwork {
			t.Error("false expected")
		}
	})
}

func TestHandler_syncDevices(t *testing.T) {
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
		mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
		handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
		syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Error("should be true")
		}
		id, ok := syncedIDs[lID]
		if !ok {
			t.Error("local device ID not in map")
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
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{
				lID: cDevice,
			})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.GetDevicesLC > 0 {
				t.Error("illegal call")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("illegal call")
			}
			if mockCC.UpdateDeviceC > 0 {
				t.Error("illegal call")
			}
		})
		t.Run("not in network equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.GetDevicesLC != 1 {
				t.Error("missing call")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("missing call")
			}
			if mockCC.UpdateDeviceC > 0 {
				t.Error("illegal call")
			}
		})
		t.Run("in network not equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			lDevice2 := lDevice
			lDevice2.Name = "test"
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice2}, map[string]models.Device{
				lID: cDevice,
			})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.GetDevicesLC > 0 {
				t.Error("illegal call")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("illegal call")
			}
			if mockCC.UpdateDeviceC != 1 {
				t.Error("missing call")
			}
			cd := mockCC.Devices[cID]
			if cd.Name != lDevice2.Name {
				t.Error("name not equal")
			}
		})
		t.Run("not in network not equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			lDevice2 := lDevice
			lDevice2.Name = "test"
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice2}, map[string]models.Device{})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
			if mockCC.GetDevicesLC != 1 {
				t.Error("missing call")
			}
			if mockCC.CreateDeviceC > 0 {
				t.Error("illegal call")
			}
			cd := mockCC.Devices[cID]
			if cd.Name != lDevice2.Name {
				t.Error("name not equal")
			}
		})
	})
	t.Run("incomplete sync", func(t *testing.T) {
		t.Run("create device error", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.CreateDeviceErr = errors.New("test")
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
			if err != nil {
				t.Error(err)
			}
			if ok {
				t.Error("should be false")
			}
			if len(syncedIDs) > 0 {
				t.Error("invalid length")
			}
		})
		t.Run("update device error", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.UpdateDeviceErr = errors.New("test")
			lDevice2 := lDevice
			lDevice2.Name = "test"
			syncedIDs, ok, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice2}, map[string]models.Device{
				lID: cDevice,
			})
			if err != nil {
				t.Error(err)
			}
			if ok {
				t.Error("should be false")
			}
			if len(syncedIDs) != 1 {
				t.Error("invalid length")
			}
		})
	})
	t.Run("request error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
		handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
		mockCC.Err = errors.New("test error")
		_, _, err := handler.syncDevices(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
		if err == nil {
			t.Error("error expected")
		}
	})
}

func TestHandler_syncDeviceIDs(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	cID := "1"
	lID := "123"
	cDevice := models.Device{
		Id:      cID,
		LocalId: lID,
	}
	lDevice := model.Device{
		ID: lID,
	}
	t.Run("cloud device does not exist", func(t *testing.T) {
		mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
		handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
		syncedIDs, ok, err := handler.syncDeviceIDs(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
		if err != nil {
			t.Error(err)
		}
		if !ok {
			t.Error("should be true")
		}
		if len(syncedIDs) > 0 {
			t.Error("invalid length")
		}
	})
	t.Run("cloud device exists", func(t *testing.T) {
		t.Run("in network", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			syncedIDs, ok, err := handler.syncDeviceIDs(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{lID: cDevice})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
		})
		t.Run("not in network", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			syncedIDs, ok, err := handler.syncDeviceIDs(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Error("should be true")
			}
			id, ok := syncedIDs[lID]
			if !ok {
				t.Error("local device ID not in map")
			}
			if id != cID {
				t.Error("cloud ID not equal")
			}
		})
		t.Run("request error", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Devices: make(map[string]models.Device), DeviceIDMap: make(map[string]string)}
			handler := &Handler{cloudClient: mockCC, attrOrigin: "test-origin"}
			mockCC.Devices[cID] = cDevice
			mockCC.DeviceIDMap[lID] = cID
			mockCC.Err = errors.New("test")
			_, _, err := handler.syncDeviceIDs(context.Background(), map[string]model.Device{lID: lDevice}, map[string]models.Device{})
			if err == nil {
				t.Error("error expected")
			}
		})
	})
}

func TestHandler_getNetwork(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("network not found", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		_, err := handler.getNetwork(context.Background())
		if err == nil {
			t.Error("error expected")
		}
		if !handler.noNetwork {
			t.Error("true expected")
		}
	})
	t.Run("network user id not equal", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		mockCC.Hubs = map[string]models.Hub{
			"1": {
				Id:        "1",
				DeviceIds: []string{"1"},
				OwnerId:   "456",
			},
		}
		handler.userID = "123"
		_, err := handler.getNetwork(context.Background())
		if err == nil {
			t.Error("error expected")
		}
		if !handler.noNetwork {
			t.Error("true expected")
		}
	})
	t.Run("request error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
			data:        data{NetworkID: "1"},
		}
		mockCC.Err = errors.New("test error")
		_, err := handler.getNetwork(context.Background())
		if err == nil {
			t.Error("error expected")
		}
		if handler.noNetwork {
			t.Error("false expected")
		}
	})
}

func TestHandler_getCloudDevs(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("no error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		cD1 := models.Device{
			Id:      "1",
			LocalId: "l1",
		}
		cD2 := models.Device{
			Id:      "2",
			LocalId: "l2",
		}
		mockCC.Devices = map[string]models.Device{
			"1": cD1,
			"2": cD2,
		}
		cloudDevices, err := handler.getCloudDevs(context.Background(), []string{"1", "2"}, mockCC.GetDevices)
		if err != nil {
			t.Error(err)
		}
		if len(cloudDevices) != 2 {
			t.Error("invalid length")
		}
		if cd := cloudDevices["l1"]; !reflect.DeepEqual(cd, cD1) {
			t.Error("not equal")
		}
		if cd := cloudDevices["l2"]; !reflect.DeepEqual(cd, cD2) {
			t.Error("not equal")
		}
	})
	t.Run("error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.Err = errors.New("test error")
		_, err := handler.getCloudDevs(context.Background(), []string{"1", "2"}, mockCC.GetDevices)
		if err == nil {
			t.Error("error expected")
		}
	})
	t.Run("device user id not equal", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		cD1 := models.Device{
			Id:      "1",
			LocalId: "l1",
			OwnerId: "123",
		}
		mockCC.Devices = map[string]models.Device{
			"1": cD1,
			"2": {
				Id:      "2",
				LocalId: "l2",
				OwnerId: "456",
			},
		}
		handler.userID = "123"
		cloudDevices, err := handler.getCloudDevs(context.Background(), []string{"1", "2"}, mockCC.GetDevices)
		if err != nil {
			t.Error(err)
		}
		if len(cloudDevices) != 1 {
			t.Error("invalid length")
		}
		if cd := cloudDevices["l1"]; !reflect.DeepEqual(cd, cD1) {
			t.Error("not equal")
		}
	})
}

func TestHandler_checkAccPols(t *testing.T) {
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("network read not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{HubsAccPol: cloud_client.HttpMethodAccPolMock{
			ReadAP:   false,
			CreateAP: true,
			UpdateAP: true,
			DeleteAP: true,
		}}
		if _, err := handler.checkAccPols(context.Background()); err == nil {
			t.Error("error expected")
		}
	})
	t.Run("network update not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{HubsAccPol: cloud_client.HttpMethodAccPolMock{
			ReadAP:   true,
			CreateAP: true,
			UpdateAP: false,
			DeleteAP: true,
		}}
		if _, err := handler.checkAccPols(context.Background()); err == nil {
			t.Error("error expected")
		}
	})
	t.Run("devices (cloud id) read not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   false,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
		}
		if _, err := handler.checkAccPols(context.Background()); err == nil {
			t.Error("error expected")
		}
	})
	t.Run("devices (local id) read not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesLAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   false,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
		}
		if _, err := handler.checkAccPols(context.Background()); err == nil {
			t.Error("error expected")
		}
	})
	t.Run("devices (cloud id) create not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: false,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesLAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
		}
		ok, err := handler.checkAccPols(context.Background())
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("false expected")
		}
	})
	t.Run("devices (cloud id) update not allowed", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: false,
				DeleteAP: true,
			},
			DevicesLAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
		}
		ok, err := handler.checkAccPols(context.Background())
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("false expected")
		}
	})
	t.Run("devices (local & cloud id) read only", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.EptAccPol = cloud_client.EndpointAccPolMock{
			HubsAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: true,
				UpdateAP: true,
				DeleteAP: true,
			},
			DevicesAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: false,
				UpdateAP: false,
				DeleteAP: false,
			},
			DevicesLAccPol: cloud_client.HttpMethodAccPolMock{
				ReadAP:   true,
				CreateAP: false,
				UpdateAP: false,
				DeleteAP: false,
			},
		}
		ok, err := handler.checkAccPols(context.Background())
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Error("false expected")
		}
	})
	t.Run("error", func(t *testing.T) {
		mockCC := &cloud_client.Mock{}
		handler := &Handler{
			cloudClient: mockCC,
		}
		mockCC.Err = errors.New("test")
		if _, err := handler.checkAccPols(context.Background()); err == nil {
			t.Error("error expected")
		}
	})
}

package cloud_hdl

import (
	"context"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/auth_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
	"time"
)

func TestHandler_Init(t *testing.T) {
	var mockCC *cloud_client.Mock
	var mockSP *auth_client.Mock
	var handler *Handler
	tmpDir := t.TempDir()
	userID := "123"
	initHandler := func() {
		mockCC = &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
		mockSP = &auth_client.Mock{UserID: userID}
		handler = &Handler{
			cloudClient:     mockCC,
			subjectProvider: mockSP,
			wrkSpacePath:    tmpDir,
		}
	}
	check := func(networkID, networkName string) {
		if handler.data.NetworkID != networkID {
			t.Error("network ID not equal")
		}
		network, ok := mockCC.Hubs[networkID]
		if !ok {
			t.Error("network ID not in map")
		}
		if network.Name != networkName {
			t.Error("network name not equal")
		}
		d, err := readData(tmpDir)
		if err != nil {
			t.Error(err)
		}
		if d.NetworkID != networkID {
			t.Error("saved network ID not equal")
		}
	}
	util.InitLogger(sb_util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("get user ID", func(t *testing.T) {
		initHandler()
		_, uID, err := handler.Init(context.Background(), "", "", time.Second)
		if err != nil {
			t.Error(err)
		}
		if uID != mockSP.UserID {
			t.Error("user id not equal")
		}
	})
	t.Run("no cloud network", func(t *testing.T) {
		networkName := "test"
		t.Run("no saved network ID parameter not set", func(t *testing.T) {
			initHandler()
			nID, _, err := handler.Init(context.Background(), "", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName)
		})
		t.Run("saved network ID parameter not set", func(t *testing.T) {
			initHandler()
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName)
		})
		t.Run("no saved networkID parameter set", func(t *testing.T) {
			initHandler()
			nID, _, err := handler.Init(context.Background(), "123", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName)
		})
		t.Run("saved networkID parameter set equal", func(t *testing.T) {
			initHandler()
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "123", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName)
		})
		t.Run("saved networkID parameter set not equal", func(t *testing.T) {
			initHandler()
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "456", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName)
		})
	})
	t.Run("cloud network exists", func(t *testing.T) {
		networkID := "134"
		network := models.Hub{
			Id:      networkID,
			Name:    "cloud",
			OwnerId: userID,
		}
		t.Run("saved network ID parameter not set", func(t *testing.T) {
			initHandler()
			mockCC.Hubs[networkID] = network
			err := writeData(tmpDir, data{NetworkID: networkID})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "", "", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID != networkID {
				t.Error("network id not equal")
			}
			check(nID, network.Name)
		})
		t.Run("saved network ID parameter set equal", func(t *testing.T) {
			initHandler()
			mockCC.Hubs[networkID] = network
			err := writeData(tmpDir, data{NetworkID: networkID})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), networkID, "", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID != networkID {
				t.Error("network id not equal")
			}
			check(nID, network.Name)
		})
		t.Run("saved network ID not equal parameter set equal", func(t *testing.T) {
			initHandler()
			mockCC.Hubs[networkID] = network
			err := writeData(tmpDir, data{NetworkID: "456"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), networkID, "", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID != networkID {
				t.Error("network id not equal")
			}
			check(nID, network.Name)
		})
		t.Run("saved network ID equal parameter set not equal", func(t *testing.T) {
			initHandler()
			mockCC.Hubs[networkID] = network
			err := writeData(tmpDir, data{NetworkID: networkID})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "456", "", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID != networkID {
				t.Error("network id not equal")
			}
			check(nID, network.Name)
		})
		t.Run("no saved networkID parameter set", func(t *testing.T) {
			initHandler()
			mockCC.Hubs[networkID] = network
			nID, _, err := handler.Init(context.Background(), networkID, "", time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, network.Name)
		})
		t.Run("user ID mismatch", func(t *testing.T) {
			initHandler()
			network.OwnerId = "456"
			mockCC.Hubs[networkID] = network
			nID, _, err := handler.Init(context.Background(), networkID, "test", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID == networkID {
				t.Error("network id is equal")
			}
			check(nID, "test")
		})
	})
}

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
	b := newCloudDevice(model.Device{
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

func Test_notEqual(t *testing.T) {
	attrOrigin := "test-origin"
	cDevice := models.Device{
		Name: "Test Device",
		Attributes: []models.Attribute{
			{
				Key:    "test-key",
				Value:  "test-val",
				Origin: attrOrigin,
			},
		},
		DeviceTypeId: "test-type",
	}
	lDevice := model.Device{
		Name: "Test Device",
		Type: "test-type",
		Attributes: []model.Attribute{
			{
				Key:   "test-key",
				Value: "test-val",
			},
		},
	}
	t.Run("cloud and local device equal", func(t *testing.T) {
		if notEqual(cDevice, lDevice, attrOrigin) {
			t.Error("should be equal")
		}
	})
	t.Run("cloud and local device name not equal", func(t *testing.T) {
		lDevice.Name = "Test Device 2"
		if !notEqual(cDevice, lDevice, attrOrigin) {
			t.Error("should not be equal")
		}
		lDevice.Name = "Test Device"
	})
	t.Run("cloud and local device type not equal", func(t *testing.T) {
		lDevice.Type = "test-type-2"
		if !notEqual(cDevice, lDevice, attrOrigin) {
			t.Error("should not be equal")
		}
		lDevice.Type = "test-type"
	})
	t.Run("cloud and local device attributes not equal", func(t *testing.T) {
		lDevice.Attributes = append(lDevice.Attributes, model.Attribute{
			Key:   "test-key-2",
			Value: "test-val-2",
		})
		if !notEqual(cDevice, lDevice, attrOrigin) {
			t.Error("should not be equal")
		}
	})
}

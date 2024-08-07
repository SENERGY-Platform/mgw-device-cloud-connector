package cloud_hdl

import (
	"context"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/auth_client"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/cloud_client"
	"github.com/SENERGY-Platform/models/go/models"
	"testing"
	"time"
)

func TestHandler_Init(t *testing.T) {
	tmpDir := t.TempDir()
	userID := "123"
	check := func(networkID, networkName string, handler *Handler, mockCC *cloud_client.Mock) {
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
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: 4})
	t.Run("get user ID", func(t *testing.T) {
		mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
		mockSP := &auth_client.Mock{UserID: userID}
		handler := &Handler{
			cloudClient:     mockCC,
			subjectProvider: mockSP,
			wrkSpacePath:    tmpDir,
		}
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
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			nID, _, err := handler.Init(context.Background(), "", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName, handler, mockCC)
		})
		t.Run("saved network ID parameter not set", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName, handler, mockCC)
		})
		t.Run("no saved networkID parameter set", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			nID, _, err := handler.Init(context.Background(), "123", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName, handler, mockCC)
		})
		t.Run("saved networkID parameter set equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "123", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName, handler, mockCC)
		})
		t.Run("saved networkID parameter set not equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			err := writeData(tmpDir, data{NetworkID: "123"})
			if err != nil {
				t.Error(err)
			}
			nID, _, err := handler.Init(context.Background(), "456", networkName, time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, networkName, handler, mockCC)
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
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
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
			check(nID, network.Name, handler, mockCC)
		})
		t.Run("saved network ID parameter set equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
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
			check(nID, network.Name, handler, mockCC)
		})
		t.Run("saved network ID not equal parameter set equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
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
			check(nID, network.Name, handler, mockCC)
		})
		t.Run("saved network ID equal parameter set not equal", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
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
			check(nID, network.Name, handler, mockCC)
		})
		t.Run("no saved networkID parameter set", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			mockCC.Hubs[networkID] = network
			nID, _, err := handler.Init(context.Background(), networkID, "", time.Second)
			if err != nil {
				t.Error(err)
			}
			check(nID, network.Name, handler, mockCC)
		})
		t.Run("user ID mismatch", func(t *testing.T) {
			mockCC := &cloud_client.Mock{Hubs: make(map[string]models.Hub)}
			mockSP := &auth_client.Mock{UserID: userID}
			handler := &Handler{
				cloudClient:     mockCC,
				subjectProvider: mockSP,
				wrkSpacePath:    tmpDir,
			}
			network.OwnerId = "456"
			mockCC.Hubs[networkID] = network
			nID, _, err := handler.Init(context.Background(), networkID, "test", time.Second)
			if err != nil {
				t.Error(err)
			}
			if nID == networkID {
				t.Error("network id is equal")
			}
			check(nID, "test", handler, mockCC)
		})
	})
}

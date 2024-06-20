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

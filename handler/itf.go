package handler

import "github.com/SENERGY-Platform/mgw-device-cloud-connector/model"

type DeviceHandler interface {
	GetDevices() map[string]model.Device
}

package dm_client

import (
	"context"
	"errors"
)

type Mock struct {
	Devices     map[string]Device
	Err         error
	GetDevicesC int
	GetDeviceC  int
}

func (m *Mock) GetDevices(_ context.Context) (map[string]Device, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	m.GetDevicesC += 1
	devices := make(map[string]Device)
	for id, device := range m.Devices {
		devices[id] = device
	}
	return devices, nil
}

func (m *Mock) GetDevice(_ context.Context, id string) (Device, error) {
	if m.Err != nil {
		return Device{}, m.Err
	}
	m.GetDeviceC += 1
	device, ok := m.Devices[id]
	if !ok {
		return Device{}, newNotFoundError(errors.New("not found"))
	}
	return device, nil
}

func (m *Mock) Reset() {
	*m = Mock{}
}

func (m *Mock) SetInternalErr() {
	m.Err = newInternalError(errors.New("internal error"))
}

func (m *Mock) SetNotFoundErr() {
	m.Err = newNotFoundError(errors.New("not found"))
}

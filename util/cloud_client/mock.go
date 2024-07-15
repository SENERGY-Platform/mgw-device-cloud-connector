package cloud_client

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/models/go/models"
	"strconv"
)

type Mock struct {
	Devices            map[string]models.Device
	Hubs               map[string]models.Hub
	DeviceIDMap        map[string]string
	AttributeOrigin    string
	EptAccPol          EndpointAccPolMock
	Err                error
	HubErr             error
	DeviceErr          error
	DevicesErr         error
	DevicesLErr        error
	AccPolErr          error
	CreateHubC         int
	GetHubC            int
	UpdateHubC         int
	CreateDeviceC      int
	GetDeviceC         int
	GetDeviceLC        int
	GetDevicesC        int
	GetDevicesLC       int
	UpdateDeviceC      int
	GetAccessPoliciesC int
}

func (m *Mock) CreateHub(_ context.Context, hub models.Hub) (string, error) {
	m.CreateHubC += 1
	if m.HubErr != nil {
		return "", m.HubErr
	}
	if m.Err != nil {
		return "", m.Err
	}
	for _, lID := range hub.DeviceLocalIds {
		rID, ok := m.DeviceIDMap[lID]
		if !ok {
			return "", newBadRequestError(errors.New("device not found"))
		}
		hub.DeviceIds = append(hub.DeviceIds, rID)
	}
	hub.Id = strconv.FormatInt(int64(len(m.Hubs))+1, 10)
	m.Hubs[hub.Id] = hub
	return hub.Id, nil
}

func (m *Mock) GetHub(_ context.Context, id string) (models.Hub, error) {
	m.GetHubC += 1
	if m.HubErr != nil {
		return models.Hub{}, m.HubErr
	}
	if m.Err != nil {
		return models.Hub{}, m.Err
	}
	hub, ok := m.Hubs[id]
	if !ok {
		return models.Hub{}, newNotFoundError(errors.New("not found"))
	}
	return hub, nil
}

func (m *Mock) UpdateHub(_ context.Context, hub models.Hub) error {
	m.UpdateHubC += 1
	if m.HubErr != nil {
		return m.HubErr
	}
	if m.Err != nil {
		return m.Err
	}
	if _, ok := m.Hubs[hub.Id]; !ok {
		return newNotFoundError(errors.New("not found"))
	}
	m.Hubs[hub.Id] = hub
	return nil
}

func (m *Mock) CreateDevice(_ context.Context, device models.Device) (string, error) {
	m.CreateDeviceC += 1
	if m.DeviceErr != nil {
		return "", m.DeviceErr
	}
	if m.Err != nil {
		return "", m.Err
	}
	if _, ok := m.DeviceIDMap[device.LocalId]; ok {
		return "", newBadRequestError(errors.New("local id exists"))
	}
	device.Id = strconv.FormatInt(int64(len(m.Devices))+1, 10)
	m.Devices[device.Id] = device
	m.DeviceIDMap[device.LocalId] = device.Id
	return device.Id, nil
}

//func (m *Mock) GetDevice(_ context.Context, id string) (models.Device, error) {
//	m.GetDeviceC += 1
//	if m.DeviceErr != nil {
//		return models.Device{}, m.DeviceErr
//	}
//	if m.Err != nil {
//		return models.Device{}, m.Err
//	}
//	device, ok := m.Devices[id]
//	if !ok {
//		return models.Device{}, newNotFoundError(errors.New("not found"))
//	}
//	return device, nil
//}
//
//func (m *Mock) GetDeviceL(_ context.Context, id string) (models.Device, error) {
//	m.GetDeviceLC += 1
//	if m.DeviceErr != nil {
//		return models.Device{}, m.DeviceErr
//	}
//	if m.Err != nil {
//		return models.Device{}, m.Err
//	}
//	device, ok := m.Devices[m.DeviceIDMap[id]]
//	if !ok {
//		return models.Device{}, newNotFoundError(errors.New("not found"))
//	}
//	return device, nil
//}

func (m *Mock) GetDevices(_ context.Context, ids []string) ([]models.Device, error) {
	m.GetDevicesC++
	if m.DevicesErr != nil {
		return nil, m.DevicesErr
	}
	if m.Err != nil {
		return nil, m.Err
	}
	var devices []models.Device
	for _, cID := range ids {
		if device, ok := m.Devices[cID]; ok {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (m *Mock) GetDevicesL(_ context.Context, ids []string) ([]models.Device, error) {
	m.GetDevicesLC++
	if m.DevicesLErr != nil {
		return nil, m.DevicesLErr
	}
	if m.Err != nil {
		return nil, m.Err
	}
	deviceMap := make(map[string]models.Device)
	for _, device := range m.Devices {
		deviceMap[device.LocalId] = device
	}
	var devices []models.Device
	for _, lID := range ids {
		if device, ok := deviceMap[lID]; ok {
			devices = append(devices, device)
		}
	}
	return devices, nil
}

func (m *Mock) UpdateDevice(_ context.Context, device models.Device, attributeOrigin string) error {
	m.AttributeOrigin = attributeOrigin
	m.UpdateDeviceC += 1
	if m.DeviceErr != nil {
		return m.DeviceErr
	}
	if m.Err != nil {
		return m.Err
	}
	if _, ok := m.Devices[device.Id]; !ok {
		return newNotFoundError(errors.New("not found"))
	}
	m.Devices[device.Id] = device
	return nil
}

func (m *Mock) GetAccessPolicies(_ context.Context) (EndpointAccPolItf, error) {
	m.GetAccessPoliciesC++
	if m.AccPolErr != nil {
		return endpointAccPol{}, m.AccPolErr
	}
	if m.Err != nil {
		return endpointAccPol{}, m.Err
	}
	return m.EptAccPol, nil
}

func (m *Mock) Reset() {
	*m = Mock{}
}

type EndpointAccPolMock struct {
	HubsAccPol     HttpMethodAccPolMock
	DevicesAccPol  HttpMethodAccPolMock
	DevicesLAccPol HttpMethodAccPolMock
}

func (m EndpointAccPolMock) Hubs() HttpMethodAccPolItf {
	return m.HubsAccPol
}

func (m EndpointAccPolMock) Devices() HttpMethodAccPolItf {
	return m.DevicesAccPol
}

func (m EndpointAccPolMock) DevicesL() HttpMethodAccPolItf {
	return m.DevicesLAccPol
}

type HttpMethodAccPolMock struct {
	ReadAP   bool
	CreateAP bool
	UpdateAP bool
	DeleteAP bool
}

func (m HttpMethodAccPolMock) Read() bool {
	return m.ReadAP
}

func (m HttpMethodAccPolMock) Create() bool {
	return m.CreateAP
}

func (m HttpMethodAccPolMock) Update() bool {
	return m.UpdateAP
}

func (m HttpMethodAccPolMock) Delete() bool {
	return m.DeleteAP
}

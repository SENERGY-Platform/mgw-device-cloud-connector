package local_device_hdl

//import (
//	"context"
//	"github.com/SENERGY-Platform/mgw-device-cloud-connector/model"
//	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/dm_client"
//	"os"
//	"reflect"
//	"testing"
//	"time"
//)
//
//func Test_newDevice(t *testing.T) {
//	a := device{
//		Device: model.Device{
//			ID:    "test",
//			Name:  "Test Device",
//			State: "online",
//			Type:  "test-type",
//			Attributes: []model.Attribute{
//				{Key: "key", Value: "val"},
//			},
//		},
//		Hash: "6733f4db65b8824b057029b33e9142486f505217",
//	}
//	dmDevice := dm_client.Device{
//		Name:     "Test Device",
//		State:    "online",
//		Type:     "test-type",
//		ModuleID: "mid",
//		Attributes: []dm_client.Attribute{
//			{Key: "key", Value: "val"},
//		},
//	}
//	b := newDevice("test", dmDevice)
//	if !reflect.DeepEqual(a, b) {
//		t.Errorf("%+v != %+v", a, b)
//	}
//	dmDevice.Name = "Test"
//	c := newDevice("test", dmDevice)
//	if a.Hash == c.Hash {
//		t.Error("hash did not change")
//	}
//}
//
//func Test_diffDevices(t *testing.T) {
//	h := New(nil, 0, 0, "")
//	hReset := func() {
//		h.devices = nil
//	}
//	devices := map[string]device{
//		"a": newDevice("a", dm_client.Device{
//			Name:  "Test Device A",
//			State: "online",
//		}),
//		"b": newDevice("b", dm_client.Device{
//			Name:  "Test Device B",
//			State: "offline",
//		}),
//		"c": newDevice("c", dm_client.Device{
//			Name:  "Test Device C",
//			State: "",
//		}),
//	}
//	t.Run("new devices", func(t *testing.T) {
//		t.Cleanup(hReset)
//		h.devices = make(map[string]device)
//		changedIDs, missingIDs, deviceMap, deviceStates := h.diffDevices(devices)
//		if len(changedIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		if len(missingIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		for id, d := range devices {
//			if _, ok := deviceMap[id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//			if d.State != "" {
//				state, ok := deviceStates[id]
//				if !ok {
//					t.Errorf("id '%s' not in state map", id)
//				}
//				if state[0] != "" {
//					t.Errorf("prev state mismatch")
//				}
//				if state[1] != d.State {
//					t.Errorf("current state mismatch")
//				}
//			} else {
//				if _, ok := deviceStates[id]; ok {
//					t.Errorf("id '%s' in state map", id)
//				}
//			}
//		}
//	})
//	t.Run("unchanged devices", func(t *testing.T) {
//		t.Cleanup(hReset)
//		h.devices = devices
//		changedIDs, missingIDs, deviceMap, deviceStates := h.diffDevices(devices)
//		if len(changedIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		if len(missingIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		if len(deviceStates) > 0 {
//			t.Error("illegal number of states")
//		}
//		for id, _ := range devices {
//			if _, ok := deviceMap[id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//		}
//	})
//	t.Run("changed devices", func(t *testing.T) {
//		t.Cleanup(hReset)
//		h.devices = devices
//		devices2 := map[string]device{
//			"a": newDevice("a", dm_client.Device{
//				Name:  "Test Device",
//				State: "online",
//			}),
//			"b": newDevice("b", dm_client.Device{
//				Name:  "Test Device B",
//				State: "online",
//			}),
//			"c": newDevice("c", dm_client.Device{
//				Name:  "Test Device C",
//				State: "offline",
//			}),
//		}
//		changedIDs, missingIDs, deviceMap, deviceStates := h.diffDevices(devices2)
//		if !inSlice("a", changedIDs) {
//			t.Error("missing id")
//		}
//		if len(missingIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		for id := range devices2 {
//			if _, ok := deviceMap[id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//		}
//		state, ok := deviceStates["b"]
//		if !ok {
//			t.Error("id not in state map")
//		}
//		if state[0] != "offline" {
//			t.Errorf("prev state mismatch")
//		}
//		if state[1] != "online" {
//			t.Errorf("current state mismatch")
//		}
//		state, ok = deviceStates["c"]
//		if !ok {
//			t.Error("id not in state map")
//		}
//		if state[0] != "" {
//			t.Errorf("prev state mismatch")
//		}
//		if state[1] != "offline" {
//			t.Errorf("current state mismatch")
//		}
//	})
//	t.Run("missing devices", func(t *testing.T) {
//		t.Cleanup(hReset)
//		h.devices = devices
//		devices2 := map[string]device{
//			"a": newDevice("a", dm_client.Device{
//				Name:  "Test Device A",
//				State: "online",
//			}),
//			"b": newDevice("b", dm_client.Device{
//				Name:  "Test Device B",
//				State: "offline",
//			}),
//		}
//		changedIDs, missingIDs, deviceMap, deviceStates := h.diffDevices(devices2)
//		if len(changedIDs) > 0 {
//			t.Error("illegal number of ids")
//		}
//		if !inSlice("c", missingIDs) {
//			t.Error("missing id")
//		}
//		if len(deviceStates) > 0 {
//			t.Error("illegal number of states")
//		}
//		for id, _ := range devices2 {
//			if _, ok := deviceMap[id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//		}
//	})
//}
//
//func Test_refreshDevices(t *testing.T) {
//	mockDMC := &dm_client.Mock{}
//	h := New(mockDMC, 0, 0, "")
//	hReset := func() {
//		h.devices = nil
//	}
//	t.Run("new, changed and missing device", func(t *testing.T) {
//		t.Cleanup(mockDMC.Reset)
//		t.Cleanup(hReset)
//		h.devices = map[string]device{
//			"a": newDevice("a", dm_client.Device{
//				Name:  "Test Device",
//				State: "offline",
//			}),
//			"b": newDevice("b", dm_client.Device{
//				Name:  "Test Device B",
//				State: "offline",
//			}),
//		}
//		mockDMC.Devices = map[string]dm_client.Device{
//			"a": {
//				Name:  "Test Device A",
//				State: "online",
//			},
//			"c": {
//				Name:  "Test Device C",
//				State: "",
//			},
//		}
//		syncCall := 0
//		h.SetDeviceSyncFunc(func(_ context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error) {
//			syncCall += 1
//			for id := range mockDMC.Devices {
//				if _, ok := devices[id]; !ok {
//					t.Errorf("id '%s' not in map", id)
//				}
//			}
//			if !inSlice("a", changedIDs) {
//				t.Error("missing id")
//			}
//			return nil, err
//		})
//		missingCall := 0
//		h.SetMissingFunc(func(_ context.Context, missingIDs []string) error {
//			missingCall += 1
//			if !inSlice("b", missingIDs) {
//				t.Error("missing id")
//			}
//			return nil
//		})
//		stateCall := 0
//		h.SetDeviceStateSyncFunc(func(_ context.Context, deviceStates map[string]string) (failed []string, err error) {
//			stateCall += 1
//			state, ok := deviceStates["a"]
//			if !ok {
//				t.Error("id not in map")
//			}
//			if state != "online" {
//				t.Error("state mismatch")
//			}
//			return nil, err
//		})
//		err := h.RefreshDevices(context.Background())
//		if err != nil {
//			t.Error(err)
//		}
//		if syncCall != 1 {
//			t.Error("illegal number of sync calls")
//		}
//		if missingCall != 1 {
//			t.Error("illegal number of missing calls")
//		}
//		if stateCall != 1 {
//			t.Error("illegal number of state change calls")
//		}
//		if len(h.devices) > 2 {
//			t.Error("illegal number of devices")
//		}
//		for id := range mockDMC.Devices {
//			if _, ok := h.devices[id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//		}
//	})
//	t.Run("failed devices", func(t *testing.T) {
//		t.Cleanup(mockDMC.Reset)
//		t.Cleanup(hReset)
//		h.devices = map[string]device{
//			"a": newDevice("a", dm_client.Device{
//				Name:  "Test Device",
//				State: "offline",
//			}),
//			"b": newDevice("b", dm_client.Device{
//				Name:  "Test Device B",
//				State: "offline",
//			}),
//		}
//		mockDMC.Devices = map[string]dm_client.Device{
//			"a": {
//				Name:  "Test Device A",
//				State: "online",
//			},
//			"c": {
//				Name:  "Test Device C",
//				State: "",
//			},
//		}
//		syncCall := 0
//		h.SetDeviceSyncFunc(func(_ context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error) {
//			syncCall += 1
//			for id := range mockDMC.Devices {
//				if _, ok := devices[id]; !ok {
//					t.Errorf("id '%s' not in map", id)
//				}
//			}
//			if !inSlice("a", changedIDs) {
//				t.Error("missing id")
//			}
//			return []string{"a", "c"}, nil
//		})
//		missingCall := 0
//		h.SetMissingFunc(func(_ context.Context, missingIDs []string) error {
//			missingCall += 1
//			if !inSlice("b", missingIDs) {
//				t.Error("missing id")
//			}
//			return nil
//		})
//		stateCall := 0
//		h.SetDeviceStateSyncFunc(func(_ context.Context, deviceStates map[string]string) (failed []string, err error) {
//			stateCall += 1
//			state, ok := deviceStates["a"]
//			if !ok {
//				t.Error("id not in map")
//			}
//			if state != "online" {
//				t.Error("state mismatch")
//			}
//			return []string{"a"}, nil
//		})
//		err := h.RefreshDevices(context.Background())
//		if err != nil {
//			t.Error(err)
//		}
//		err = h.RefreshDevices(context.Background())
//		if err != nil {
//			t.Error(err)
//		}
//		if syncCall != 2 {
//			t.Error("illegal number of sync calls")
//		}
//		if missingCall != 1 {
//			t.Error("illegal number of missing calls")
//		}
//		if stateCall != 2 {
//			t.Error("illegal number of state change calls")
//		}
//		if len(h.devices) != 1 {
//			t.Error("illegal number of devices")
//		}
//		if _, ok := h.devices["a"]; !ok {
//			t.Error("id not in map")
//		}
//	})
//	t.Run("error case", func(t *testing.T) {
//		t.Cleanup(mockDMC.Reset)
//		t.Cleanup(hReset)
//		mockDMC.SetInternalErr()
//		err := h.RefreshDevices(context.Background())
//		if err == nil {
//			t.Error(err)
//		}
//	})
//}
//
//func Test_refreshDevicesWithPrefix(t *testing.T) {
//	mockDMC := &dm_client.Mock{}
//	h := New(mockDMC, 0, 0, "prefix-")
//	hReset := func() {
//		h.devices = nil
//	}
//	t.Run("new, changed and missing device", func(t *testing.T) {
//		t.Cleanup(mockDMC.Reset)
//		t.Cleanup(hReset)
//		h.devices = map[string]device{
//			"prefix-a": newDevice("prefix-a", dm_client.Device{
//				Name:  "Test Device",
//				State: "offline",
//			}),
//			"prefix-b": newDevice("prefix-b", dm_client.Device{
//				Name:  "Test Device B",
//				State: "offline",
//			}),
//		}
//		mockDMC.Devices = map[string]dm_client.Device{
//			"a": {
//				Name:  "Test Device A",
//				State: "online",
//			},
//			"c": {
//				Name:  "Test Device C",
//				State: "",
//			},
//		}
//		syncCall := 0
//		h.SetDeviceSyncFunc(func(_ context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error) {
//			syncCall += 1
//			for id := range mockDMC.Devices {
//				if _, ok := devices["prefix-"+id]; !ok {
//					t.Errorf("id '%s' not in map", id)
//				}
//			}
//			if !inSlice("prefix-a", changedIDs) {
//				t.Error("missing id")
//			}
//
//			return nil, err
//		})
//		missingCall := 0
//		h.SetMissingFunc(func(_ context.Context, missingIDs []string) error {
//			missingCall += 1
//			if !inSlice("prefix-b", missingIDs) {
//				t.Error("missing id")
//			}
//			return nil
//		})
//		stateCall := 0
//		h.SetDeviceStateSyncFunc(func(_ context.Context, deviceStates map[string]string) (failed []string, err error) {
//			stateCall += 1
//			state, ok := deviceStates["prefix-a"]
//			if !ok {
//				t.Error("id not in map")
//			}
//			if state != "online" {
//				t.Error("state mismatch")
//			}
//			return nil, err
//		})
//		err := h.RefreshDevices(context.Background())
//		if err != nil {
//			t.Error(err)
//		}
//		if syncCall != 1 {
//			t.Error("illegal number of sync calls")
//		}
//		if missingCall != 1 {
//			t.Error("illegal number of missing calls")
//		}
//		if stateCall != 1 {
//			t.Error("illegal number of state change calls")
//		}
//		if len(h.devices) > 2 {
//			t.Error("illegal number of devices")
//		}
//		for id := range mockDMC.Devices {
//			if _, ok := h.devices["prefix-"+id]; !ok {
//				t.Errorf("id '%s' not in map", id)
//			}
//		}
//	})
//}
//
//func Test_loop(t *testing.T) {
//	mockDMC := &dm_client.Mock{
//		Devices: map[string]dm_client.Device{
//			"test": {
//				Name:  "Test Device",
//				State: "online",
//			},
//		},
//	}
//	h := New(mockDMC, 0, time.Millisecond*100, "")
//	syncCall := 0
//	h.SetDeviceSyncFunc(func(_ context.Context, devices map[string]model.Device, changedIDs []string) (failed []string, err error) {
//		syncCall += 1
//		return nil, err
//	})
//	h.Start()
//	time.Sleep(time.Millisecond * 500)
//	h.Stop()
//	ticker := time.NewTicker(time.Millisecond * 100)
//	timer := time.NewTimer(time.Millisecond * 500)
//	defer ticker.Stop()
//	defer func() {
//		if !timer.Stop() {
//			<-timer.C
//		}
//	}()
//	stop := false
//	for !stop {
//		select {
//		case <-timer.C:
//			t.Error("timeout occurred")
//			os.Exit(1)
//		case <-ticker.C:
//			if !h.running {
//				stop = true
//				break
//			}
//		}
//	}
//	if syncCall == 0 {
//		t.Error("no calls")
//	}
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

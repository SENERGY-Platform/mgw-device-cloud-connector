/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package debounce

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/y-du/go-log-level/level"
)

type testCloudDeviceProvider struct {
	devices map[string]models.Device
}

func (m *testCloudDeviceProvider) GetCloudDevice(localId string) *models.Device {
	d, ok := m.devices[localId]
	if !ok {
		return nil
	}
	return &d
}

type testMessageRelayHandler struct {
	c chan handler.Message
}

func (m *testMessageRelayHandler) Put(msg handler.Message) error {
	m.c <- msg
	return nil
}

func TestDebouncer(t *testing.T) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: level.Debug})

	testCloudDeviceProvider, testRelayHandler := setup()

	debouncer, err := New(testCloudDeviceProvider, testRelayHandler, util.DebouncerConfig{RefreshDebounceConfigInterval: "1ms"}, t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test No Debounce for unknown device", func(t *testing.T) {
		start := time.Now()
		err := debouncer.Put(util.Message{
			TopicVal:   "event/unknown_device/service1",
			PayloadVal: "test1",
		})
		if err != nil {
			t.Fatal(err)
		}
		<-testRelayHandler.c
		elapsed := time.Since(start)
		if elapsed > 10*time.Millisecond {
			t.Fatal("message was delayed even though no debounce should be applied")
		}
	})

	t.Run("Test No Debounce for service with no debounce attribute", func(t *testing.T) {
		start := time.Now()
		err := debouncer.Put(util.Message{
			TopicVal:   "event/device1/service2",
			PayloadVal: "test1",
		})
		if err != nil {
			t.Fatal(err)
		}
		<-testRelayHandler.c
		elapsed := time.Since(start)
		if elapsed > 10*time.Millisecond {
			t.Fatal("message was delayed even though no debounce should be applied")
		}
	})

	t.Run("Test Debounce for service with debounce attribute", func(t *testing.T) {
		var start time.Time
		var m handler.Message
		t.Run("Test no debounce for first message", func(t *testing.T) {
			start = time.Now()
			err := debouncer.Put(util.Message{
				TopicVal:   "event/device1/service1",
				PayloadVal: "test1",
				TimeVal:    start,
			})
			if err != nil {
				t.Fatal(err)
			}
			<-testRelayHandler.c
			elapsed := time.Since(start)
			if elapsed > 10*time.Millisecond {
				// No debounce for first message
				t.Fatal("message was delayed even though no debounce should be applied")
			}
		})
		t.Run("Test debounce for successive messages", func(t *testing.T) {
			start = time.Now()
			err := debouncer.Put(util.Message{
				TopicVal:   "event/device1/service1",
				PayloadVal: "test2",
				TimeVal:    start,
			})
			if err != nil {
				t.Fatal(err)
			}
			err = debouncer.Put(util.Message{
				TopicVal:   "event/device1/service1",
				PayloadVal: "test3",
				TimeVal:    start,
			})
			if err != nil {
				t.Fatal(err)
			}
			m = <-testRelayHandler.c
			elapsed := time.Since(start)
			if elapsed < 100*time.Millisecond {
				t.Fatal("message was not delayed even though debounce should be applied")
			}
			if string(m.Payload()) != "test3" {
				t.Fatal("debounced message payload invalid")
			}
		})

		t.Run("Test timestamp unchanged after debounce", func(t *testing.T) {
			if !m.Timestamp().Equal(start) {
				t.Fatal("message timestamp was modified during debounce")
			}
		})

		t.Run("Test debounce update interval", func(t *testing.T) {
			testCloudDeviceProvider.devices["device1"] = models.Device{
				Id:      "cloud-device-1",
				LocalId: "device1",
				Attributes: []models.Attribute{
					{
						Key:   "debounce/service1",
						Value: "200ms",
					},
				},
			}
			time.Sleep(5 * time.Millisecond)
			start = time.Now()
			err := debouncer.Put(util.Message{
				TopicVal:   "event/device1/service1",
				PayloadVal: "test4",
				TimeVal:    start,
			})
			if err != nil {
				t.Fatal(err)
			}
			m = <-testRelayHandler.c
			err = debouncer.Put(util.Message{
				TopicVal:   "event/device1/service1",
				PayloadVal: "test5",
				TimeVal:    start,
			})
			if err != nil {
				t.Fatal(err)
			}

			m = <-testRelayHandler.c
			elapsed := time.Since(start)
			if elapsed < 200*time.Millisecond {
				t.Fatal("message was not delayed even though debounce should be applied")
			}
			if string(m.Payload()) != "test5" {
				t.Fatal("debounced message payload invalid")
			}

		})
	})
}

func TestDebouncerInMemoryStorage(t *testing.T) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: level.Debug})

	testCloudDeviceProvider, testRelayHandler := setup()

	debouncer, err := New(testCloudDeviceProvider, testRelayHandler, util.DebouncerConfig{RefreshDebounceConfigInterval: "1ms"}, t.Context())
	if err != nil {
		t.Fatal(err)
	}

	err = testDebouncerStorageProvider(debouncer, testRelayHandler)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDebouncerSQLiteStorage(t *testing.T) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: level.Debug})

	testCloudDeviceProvider, testRelayHandler := setup()

	err := os.Remove("/tmp/debounce.sqlite")
	if err != nil && !errors.Is(err, syscall.ENOENT) {
		t.Fatal(err)
	}
	defer os.Remove("/tmp/debounce.sqlite")

	ctx, cancel := context.WithCancel(t.Context())

	debouncer, err := New(testCloudDeviceProvider, testRelayHandler, util.DebouncerConfig{RefreshDebounceConfigInterval: "1ms", PersistencePath: "/tmp/debounce.sqlite"}, ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Test SQLiteStorage", func(t *testing.T) {
		err = testDebouncerStorageProvider(debouncer, testRelayHandler)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Test Debouncer starts workers from storage", func(t *testing.T) {
		cancel()
		time.Sleep(time.Second)

		d2, err := New(debouncer.cloudDeviceProvider, debouncer.child, debouncer.config, t.Context())
		if err != nil {
			t.Fatal(err)
		}

		workers := slices.Collect(maps.Values(d2.workers))
		if len(workers) != 1 {
			t.Fatal(fmt.Errorf("debouncer did not start workers"))
		}
	})
}

func testDebouncerStorageProvider(debouncer *Debouncer, testRelayHandler *testMessageRelayHandler) error {
	start := time.Now()

	// overwrite storage provider for test sync
	testWrappedStorageProvider := &testWrappedStorageProvider{
		storageProvider: debouncer.storageProvider,
		stored:          make(chan bool),
	}
	debouncer.storageProvider = testWrappedStorageProvider

	err := debouncer.Put(util.Message{
		TopicVal:   "event/device1/service1",
		PayloadVal: "test1",
		TimeVal:    start,
	})
	if err != nil {
		return err
	}
	<-testWrappedStorageProvider.stored
	<-testRelayHandler.c

	states, err := debouncer.storageProvider.getAllStates()
	if err != nil {
		return err
	}

	if len(states) != 1 {
		return fmt.Errorf("message state not stored")
	}
	if states[0].BufferedMsg != nil {
		return fmt.Errorf("stored message state payload incorrect")
	}
	if states[0].NextForward.Before(start) {
		return fmt.Errorf("stored message state nextForward incorrect")
	}

	err = debouncer.Put(util.Message{
		TopicVal:   "event/device1/service1",
		PayloadVal: "test2",
		TimeVal:    time.Now(),
	})
	if err != nil {
		return err
	}
	<-testWrappedStorageProvider.stored

	err = debouncer.Put(util.Message{
		TopicVal:   "event/device1/service1",
		PayloadVal: "test3",
		TimeVal:    time.Now(),
	})
	if err != nil {
		return err
	}
	<-testWrappedStorageProvider.stored

	states, err = debouncer.storageProvider.getAllStates()
	if err != nil {
		return err
	}
	if states[0].BufferedMsg == nil {
		return fmt.Errorf("stored message state payload incorrect")
	}

	msg := states[0].BufferedMsg
	if string(msg.Payload()) != "test3" {
		return fmt.Errorf("stored message state payload incorrect")
	}
	if states[0].NextForward.Before(start) {
		return fmt.Errorf("stored message state nextForward incorrect")
	}
	if states[0].Id != "event/device1/service1" {
		return fmt.Errorf("stored message state id incorrect")
	}

	return nil
}

func setup() (*testCloudDeviceProvider, *testMessageRelayHandler) {
	util.InitLogger(util.LoggerConfig{Terminal: true, Level: level.Debug})

	testCloudDeviceProvider := &testCloudDeviceProvider{
		devices: map[string]models.Device{
			"device1": {
				Id:      "cloud-device-1",
				LocalId: "device1",
				Attributes: []models.Attribute{
					{
						Key:   "debounce/service1",
						Value: "100ms",
					},
				},
			},
		},
	}

	testRelayHandler := &testMessageRelayHandler{
		c: make(chan handler.Message, 1),
	}

	return testCloudDeviceProvider, testRelayHandler
}

type testWrappedStorageProvider struct {
	stored          chan bool
	storageProvider storageProvider
}

func (t *testWrappedStorageProvider) storeState(state WorkerState) error {
	defer func() {
		if t.stored != nil {
			t.stored <- true
		}
	}()
	return t.storageProvider.storeState(state)
}

func (t *testWrappedStorageProvider) getAllStates() ([]WorkerState, error) {
	return t.storageProvider.getAllStates()
}

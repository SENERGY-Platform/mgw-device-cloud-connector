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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/models/go/models"
)

type CloudDeviceProvider interface {
	GetCloudDevice(localId string) *models.Device
}

type Debouncer struct {
	cloudDeviceProvider CloudDeviceProvider
	child               handler.MessageRelayHandler
	config              util.DebouncerConfig
	workers             map[string]*worker
	ctx                 context.Context
	mux                 sync.RWMutex
	updateInterval      time.Duration
	storageProvider     storageProvider
}

func New(cloudDeviceProvider CloudDeviceProvider, child handler.MessageRelayHandler, config util.DebouncerConfig, ctx context.Context) (*Debouncer, error) {
	var updateInterval time.Duration
	if config.RefreshDebounceConfigInterval == "" {
		updateInterval = 1 * time.Minute
	} else {
		var err error
		updateInterval, err = time.ParseDuration(config.RefreshDebounceConfigInterval)
		if err != nil {
			return nil, err
		}
	}
	var storageProvider storageProvider
	var err error
	if config.PersistencePath == "" {
		storageProvider, err = newInMemoryStorageProvider()
	} else {
		storageProvider, err = newSqliteStorageProvider(config.PersistencePath)
	}
	if err != nil {
		return nil, err
	}
	d := &Debouncer{
		cloudDeviceProvider: cloudDeviceProvider,
		child:               child,
		config:              config,
		workers:             make(map[string]*worker),
		ctx:                 ctx,
		updateInterval:      updateInterval,
		storageProvider:     storageProvider,
	}
	err = d.startWorkersFromStorage()
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Debouncer) Put(m handler.Message) error {
	util.Logger.Debug("debouncer: put called")
	d.mux.RLock()
	w, ok := d.workers[m.Topic()]
	d.mux.RUnlock()
	if !ok {
		topicsParts := strings.Split(m.Topic(), "/")
		if len(topicsParts) != 3 {
			// valid topic formats are event/{device_id}/{service_id}
			// slashes in device_id or service_id are not allowed
			return fmt.Errorf("invalid topic format: %s", m.Topic())
		}
		device := d.cloudDeviceProvider.GetCloudDevice(topicsParts[1])
		if device == nil {
			util.Logger.Debugf("debouncer: no cloud device found for topic %s", m.Topic())
			return d.child.Put(m)
		}
		debounceDuration, err := d.getDebounceDuration(m)
		if err != nil {
			return err
		}
		if debounceDuration == nil {
			util.Logger.Debugf("debouncer: no debounce duration configured for topic %s", m.Topic())
			return d.child.Put(m)
		}
		w = newWorker(m.Topic(), d.child, d.ctx, *debounceDuration, time.Time{}, d.storageProvider)
		d.mux.Lock()
		d.workers[m.Topic()] = w
		d.mux.Unlock()
	} else if time.Now().After(w.debounceForUpdatedAt.Add(d.updateInterval)) {
		// update debounce duration to reflect cloud config changes
		debounceDuration, err := d.getDebounceDuration(m)
		if err != nil {
			return err
		}
		w.debounceFor = *debounceDuration
		w.debounceForUpdatedAt = time.Now()
	}
	return w.put(m)
}

func (d *Debouncer) startWorkersFromStorage() error {
	states, err := d.storageProvider.getAllStates()
	if err != nil {
		return err
	}
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, state := range states {
		m := util.Message{
			TopicVal: state.Id,
		}
		dur, err := d.getDebounceDuration(m)
		if err != nil {
			util.Logger.Warningf("debouncer: load worker from storage for topic %s: %s", state.Id, err)
		}
		if dur == nil {
			var d = time.Duration(0)
			dur = &d
		}
		w := newWorker(state.Id, d.child, d.ctx, *dur, state.NextForward, d.storageProvider)
		d.workers[state.Id] = w
	}
	return nil
}

// returns the debounce duration configured for the given device or nil if not configured or unable to parse configured value.
func (d *Debouncer) getDebounceDuration(m handler.Message) (*time.Duration, error) {
	topicsParts := strings.Split(m.Topic(), "/")
	if len(topicsParts) != 3 {
		// valid topic formats are event/{device_id}/{service_id}
		// slashes in device_id or service_id are not allowed
		return nil, fmt.Errorf("invalid topic format: %s", m.Topic())
	}
	device := d.cloudDeviceProvider.GetCloudDevice(topicsParts[1])
	if device == nil {
		util.Logger.Debugf("debouncer: no cloud device found for topic %s", m.Topic())
		return nil, nil
	}

	for _, attr := range device.Attributes {
		if attr.Key != "debounce/"+topicsParts[2] {
			continue
		}

		dur, err := time.ParseDuration(attr.Value)
		if err != nil {
			util.Logger.Warningf("debounce: parse debounce duration for device %s, localServiceId %s: %s", device.Id, topicsParts[2], err)
			return nil, nil
		}
		return &dur, nil
	}
	return nil, nil
}

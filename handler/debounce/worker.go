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
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/handler"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
)

type worker struct {
	ctx                  context.Context
	child                handler.MessageRelayHandler
	input                chan handler.Message
	debounceFor          time.Duration
	debounceForUpdatedAt time.Time
	state                WorkerState
	storageProvider      storageProvider
}

type WorkerState struct {
	NextForward time.Time
	BufferedMsg *util.Message // json Marshal does not work on interface with getters methods
	Id          string
}

func newWorker(id string, child handler.MessageRelayHandler, ctx context.Context, debounceFor time.Duration, nextForward time.Time, storageProvider storageProvider) *worker {
	w := &worker{
		state: WorkerState{
			Id:          id,
			NextForward: nextForward,
		},
		child:                child,
		ctx:                  ctx,
		input:                make(chan handler.Message),
		debounceFor:          debounceFor,
		debounceForUpdatedAt: time.Now(),
		storageProvider:      storageProvider,
	}
	go w.run()
	return w
}

func (w *worker) put(m handler.Message) error {
	util.Logger.Debug("debounce worker: put called")
	w.input <- m
	return nil
}

func (w *worker) run() {
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case m := <-w.input:
			if w.state.BufferedMsg == nil {
				// currently not waiting to debounce
				// check if debounce required
				if time.Now().After(w.state.NextForward) {
					// debounce period already passed, forward immediately
					w.forward(m)
					continue
				}
				// debounce required, start timer
				timer.Reset(time.Until(w.state.NextForward))
				msg := util.MessageFromHandler(m)
				w.state.BufferedMsg = &msg
				w.storeState()
			} else {
				// already waiting to debounce, replace stored message
				msg := util.MessageFromHandler(m)
				w.state.BufferedMsg = &msg
				w.storeState()
			}
		case <-timer.C:
			if w.state.BufferedMsg == nil {
				util.Logger.Warning("debounce worker: timer fired but no message to forward, implementation error!")
				continue
			}
			w.forward(w.state.BufferedMsg)
			continue
		}
	}
}

func (w *worker) forward(msg handler.Message) {
	util.Logger.Debug("debounce worker: forwarding message")
	err := w.child.Put(msg)
	if err != nil {
		util.Logger.Errorf("debounce worker: relay message: %s", err)
		return
	}
	w.state.NextForward = msg.Timestamp().Add(w.debounceFor)
	w.state.BufferedMsg = nil
	w.storeState()
}

func (w *worker) storeState() {
	util.Logger.Debugf("debounce worker: storing state, state has message %t", w.state.BufferedMsg != nil)
	err := w.storageProvider.storeState(w.state)
	if err != nil {
		util.Logger.Errorf("debounce worker: store state: %s", err)
		return
	}
}

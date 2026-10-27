/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package health_hdl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util"
	"github.com/SENERGY-Platform/mgw-device-cloud-connector/util/topic"
)

const LogPrefix = "[health-hdl]"

const (
	maxBodySize     = 1 << 20
	modulesPath     = "/health/deployments?auxiliary_deployments=true"
	corePath        = "/health"
	maxErrorTextLen = 200
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type published struct {
	state int
	keys  []string
	at    time.Time
}

type Handler struct {
	httpClient        httpClient
	mmBaseUrl         string
	cmBaseUrl         string
	publishFunc       func(topic string, data []byte) error
	checkInterval     time.Duration
	heartbeatInterval time.Duration
	runningTimeout    time.Duration
	now               func() time.Time
	warnf             func(format string, args ...any)
	last              *published
	warnedState       *int
	ctx               context.Context
	cf                context.CancelFunc
	wg                sync.WaitGroup
}

func New(httpClient httpClient, mmBaseUrl, cmBaseUrl string, publishFunc func(topic string, data []byte) error, checkInterval, heartbeatInterval, updateRunningTimeout time.Duration) (*Handler, error) {
	if checkInterval <= 0 || heartbeatInterval <= 0 || updateRunningTimeout <= 0 {
		return nil, fmt.Errorf("invalid intervals: check %s, heartbeat %s, update running timeout %s", checkInterval, heartbeatInterval, updateRunningTimeout)
	}
	ctx, cf := context.WithCancel(context.Background())
	return &Handler{
		httpClient:        httpClient,
		mmBaseUrl:         strings.TrimRight(mmBaseUrl, "/"),
		cmBaseUrl:         strings.TrimRight(cmBaseUrl, "/"),
		publishFunc:       publishFunc,
		checkInterval:     checkInterval,
		heartbeatInterval: heartbeatInterval,
		runningTimeout:    updateRunningTimeout,
		now:               time.Now,
		warnf:             func(format string, args ...any) { util.Logger.Warningf(format, args...) },
		ctx:               ctx,
		cf:                cf,
	}, nil
}

func (h *Handler) Start() {
	h.wg.Add(1)
	go h.run()
}

func (h *Handler) Stop() {
	h.cf()
	h.wg.Wait()
}

func (h *Handler) run() {
	defer h.wg.Done()
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()
	for {
		h.check(h.ctx)
		select {
		case <-ticker.C:
		case <-h.ctx.Done():
			return
		}
	}
}

// check reads both sources, derives the state and publishes the report if it is due.
func (h *Handler) check(ctx context.Context) {
	now := h.now().UTC()
	modulesRaw, modules, modulesErr := h.readModules(ctx)
	coreRaw, core, coreErr, coreUnavailable := h.readCore(ctx)
	state, reasons, keys := deriveHealth(modules, modulesErr != "", core, coreErr != "" && !coreUnavailable, now, h.runningTimeout)
	if !h.publishDue(state, keys, now) {
		return
	}
	data, err := marshalReport(Report{
		Version:      reportVersion,
		Time:         now.Format(time.RFC3339),
		State:        state,
		Reasons:      reasons,
		Core:         coreRaw,
		CoreError:    coreErr,
		Modules:      modulesRaw,
		ModulesError: modulesErr,
	})
	if err != nil {
		util.Logger.Errorf("%s marshal report: %s", LogPrefix, err)
		return
	}
	if err = h.publishFunc(topic.CloudHealthPub, data); err != nil {
		if h.warnedState == nil || *h.warnedState != state {
			h.warnf("%s publish report (state=%d): %s", LogPrefix, state, err)
			h.warnedState = &state
		}
		return
	}
	h.warnedState = nil
	if h.last == nil || h.last.state != state {
		util.Logger.Infof("%s published report (state=%d, reasons=%d)", LogPrefix, state, len(reasons))
	}
	h.last = &published{state: state, keys: keys, at: now}
}

func (h *Handler) publishDue(state int, keys []string, now time.Time) bool {
	if h.last == nil || h.last.state != state || !slices.Equal(h.last.keys, keys) {
		return true
	}
	return now.Sub(h.last.at) >= h.heartbeatInterval
}

func (h *Handler) readModules(ctx context.Context) (json.RawMessage, *modulesInfo, string) {
	body, status, err := h.get(ctx, h.mmBaseUrl+modulesPath)
	if err != nil {
		return nil, nil, errorText(err.Error())
	}
	if status != http.StatusOK {
		return nil, nil, fmt.Sprintf("unexpected status %d", status)
	}
	var info modulesInfo
	if err = json.Unmarshal(body, &info); err != nil {
		return nil, nil, errorText("invalid response: " + err.Error())
	}
	return body, &info, ""
}

// readCore reports unavailable for 404 and 403, which an installer without the health route answers.
func (h *Handler) readCore(ctx context.Context) (raw json.RawMessage, info *coreInfo, errTxt string, unavailable bool) {
	body, status, err := h.get(ctx, h.cmBaseUrl+corePath)
	if err != nil {
		return nil, nil, errorText(err.Error()), false
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotFound, http.StatusForbidden:
		return nil, nil, fmt.Sprintf("core health not available (status %d)", status), true
	default:
		return nil, nil, fmt.Sprintf("unexpected status %d", status), false
	}
	var ci coreInfo
	if err = json.Unmarshal(body, &ci); err != nil {
		return nil, nil, errorText("invalid response: " + err.Error()), false
	}
	return body, &ci, "", false
}

func (h *Handler) get(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize+1))
	if err != nil {
		return nil, 0, err
	}
	if len(body) > maxBodySize {
		return nil, resp.StatusCode, errors.New("response too large")
	}
	return body, resp.StatusCode, nil
}

func errorText(s string) string {
	return truncate(s, maxErrorTextLen)
}
